package main

import (
	"context"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/dangvu/zetipay/internal/config"
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gorilla/websocket"
	"github.com/joho/godotenv"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: Restrict to your frontend domain in production
	},
}

type TransactionEvent struct {
	TxHash       string `json:"txHash"`
	From         string `json:"from"`
	To           string `json:"to"`
	Value        string `json:"value"`
	TokenAddress string `json:"tokenAddress,omitempty"`
}

const erc20ABI = `[
	{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}
]`

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize Ethereum client
	client, err := ethclient.Dial(cfg.Web3ProviderURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	// Handle WebSocket connections
	http.HandleFunc("/ws/transactions/", func(w http.ResponseWriter, r *http.Request) {
		address := r.URL.Path[len("/ws/transactions/"):]
		if !common.IsHexAddress(address) {
			http.Error(w, "Invalid wallet address", http.StatusBadRequest)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// Monitor transactions and balance
		go monitorTransactions(client, address, conn)

		<-make(chan struct{}) // Keep connection open
	})

	// Start WebSocket server
	log.Printf("WebSocket server running on port %s", cfg.WSPort)
	if err := http.ListenAndServe(":"+cfg.WSPort, nil); err != nil {
		log.Fatalf("Failed to start WebSocket server: %v", err)
	}
}

func monitorTransactions(client *ethclient.Client, address string, conn *websocket.Conn) {
	ctx := context.Background()
	addr := common.HexToAddress(address)

	// Subscribe to token transfers
	tokenAddresses := []string{
		"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", // UNI (Sepolia)
		// Add more tokens as needed
	}
	for _, tokenAddr := range tokenAddresses {
		go func(tokenAddr string) {
			for {
				subscribeToTokenTransfers(client, address, tokenAddr, conn)
				log.Printf("Retrying subscription for token %s in 10 seconds", tokenAddr)
				time.Sleep(10 * time.Second)
			}
		}(tokenAddr)
	}

	// Periodically check ETH balance
	for {
		balance, err := client.BalanceAt(ctx, addr, nil)
		if err != nil {
			log.Printf("Failed to fetch balance for %s: %v", address, err)
			if err := conn.WriteJSON(map[string]string{
				"event": "error",
				"error": fmt.Sprintf("Failed to fetch balance: %v", err),
			}); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
			time.Sleep(10 * time.Second)
			continue
		}
		balanceFloat := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))
		err = conn.WriteJSON(map[string]interface{}{
			"event":      "balance_update",
			"address":    address,
			"ethBalance": balanceFloat.Text('f', 6),
		})
		if err != nil {
			log.Printf("WebSocket write error: %v", err)
			return
		}
		time.Sleep(30 * time.Second)
	}
}

func subscribeToTokenTransfers(client *ethclient.Client, address, tokenAddress string, conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tokenAddr := common.HexToAddress(tokenAddress)
	addr := common.HexToAddress(address)

	// Get token decimals
	parsed, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		log.Printf("Failed to parse ERC-20 ABI for %s: %v", tokenAddress, err)
		return
	}
	dataDecimals, err := parsed.Pack("decimals")
	if err != nil {
		log.Printf("Failed to pack decimals call for %s: %v", tokenAddress, err)
		return
	}
	resultDecimals, err := client.CallContract(context.Background(), ethereum.CallMsg{
		To:   &tokenAddr,
		Data: dataDecimals,
	}, nil)
	if err != nil {
		log.Printf("Failed to call decimals for %s: %v", tokenAddress, err)
		return
	}
	var decimals uint8
	if err := parsed.UnpackIntoInterface(&decimals, "decimals", resultDecimals); err != nil {
		log.Printf("Failed to unpack decimals for %s: %v", tokenAddress, err)
		return
	}
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)

	// Transfer event signature
	transferSig := []byte("Transfer(address,address,uint256)")
	transferTopic := crypto.Keccak256Hash(transferSig)

	// Filter for outgoing transfers (from == address)
	queryFrom := ethereum.FilterQuery{
		Addresses: []common.Address{tokenAddr},
		Topics: [][]common.Hash{
			{transferTopic},
			{common.BytesToHash(addr.Bytes())},
		},
	}

	// Filter for incoming transfers (to == address)
	queryTo := ethereum.FilterQuery{
		Addresses: []common.Address{tokenAddr},
		Topics: [][]common.Hash{
			{transferTopic},
			{},
			{common.BytesToHash(addr.Bytes())},
		},
	}

	// Subscribe to logs
	logs := make(chan types.Log)
	subFrom, err := client.SubscribeFilterLogs(ctx, queryFrom, logs)
	if err != nil {
		log.Printf("Failed to subscribe to logs (from) for %s: %v", tokenAddress, err)
		return
	}
	defer subFrom.Unsubscribe()

	subTo, err := client.SubscribeFilterLogs(ctx, queryTo, logs)
	if err != nil {
		log.Printf("Failed to subscribe to logs (to) for %s: %v", tokenAddress, err)
		return
	}
	defer subTo.Unsubscribe()

	for {
		select {
		case err := <-subFrom.Err():
			log.Printf("Subscription error (from) for %s: %v", tokenAddress, err)
			return
		case err := <-subTo.Err():
			log.Printf("Subscription error (to) for %s: %v", tokenAddress, err)
			return
		case vLog := <-logs:
			var fromAddr, toAddr common.Address
			if len(vLog.Topics) > 1 {
				fromAddr = common.BytesToAddress(vLog.Topics[1].Bytes())
			}
			if len(vLog.Topics) > 2 {
				toAddr = common.BytesToAddress(vLog.Topics[2].Bytes())
			}
			value := new(big.Int).SetBytes(vLog.Data)
			valueFloat := new(big.Float).Quo(new(big.Float).SetInt(value), new(big.Float).SetInt(denominator))

			event := TransactionEvent{
				TxHash:       vLog.TxHash.Hex(),
				From:         fromAddr.Hex(),
				To:           toAddr.Hex(),
				Value:        valueFloat.Text('f', 6),
				TokenAddress: tokenAddress,
			}
			if err := conn.WriteJSON(event); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}