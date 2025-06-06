package blockchain

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Transaction represents an Ethereum transaction
type Transaction struct {
	TxHash    string `json:"txHash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Amount    string `json:"amount"`
	Timestamp uint64 `json:"timestamp"`
}

type EtherscanTx struct {
	Hash      string `json:"hash"`
	From      string `json:"from"`
	To        string `json:"to"`
	Value     string `json:"value"`
	GasUsed   string `json:"gasUsed"`
	TimeStamp string `json:"timeStamp"`
}

type EtherscanResponse struct {
	Status  string        `json:"status"`
	Message string        `json:"message"`
	Result  []EtherscanTx `json:"result"`
}

func GetBalances(client *ethclient.Client, address string) (string, map[string]string, error) {
	addr := common.HexToAddress(address)
	ctx := context.Background()

	// Get ETH balance
	ethBalance, err := client.BalanceAt(ctx, addr, nil)
	if err != nil {
		return "", nil, err
	}
	ethBalanceFloat := new(big.Float).Quo(new(big.Float).SetInt(ethBalance), big.NewFloat(1e18))

	// Get token balances
	tokenAddresses := []string{
		"0x1f9840a85d5aF5bf1D1762F925BDADdC4201F984", // UNI (Sepolia example)
		"0xA0b86991c6218b36c1d19D4a2e9Eb0cE3606eB48", // USDC (mainnet, replace for Sepolia)
	}
	tokenBalances := make(map[string]string)
	for _, tokenAddr := range tokenAddresses {
		balance, err := GetTokenBalance(client, address, tokenAddr)
		if err != nil {
			continue
		}
		tokenBalances[tokenAddr] = balance
	}

	return ethBalanceFloat.Text('f', 6), tokenBalances, nil
}

func SendETHTransaction(client *ethclient.Client, fromAddress, toAddress, amount, privateKey string) (string, error) {
	ctx := context.Background()
	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)

	amountFloat, ok := new(big.Float).SetString(amount)
	if !ok {
		return "", fmt.Errorf("invalid amount format")
	}
	wei := new(big.Int)
	amountFloat.Mul(amountFloat, big.NewFloat(1e18)).Int(wei)

	// Placeholder for signature-based flow
	// In production, signature is verified in middleware
	if privateKey == "" {
		return "", fmt.Errorf("private key required for testing; use signature in production")
	}

	privKey, err := crypto.HexToECDSA(strings.TrimPrefix(privateKey, "0x"))
	if err != nil {
		return "", fmt.Errorf("invalid private key: %v", err)
	}

	nonce, err := client.PendingNonceAt(ctx, fromAddr)
	if err != nil {
		return "", err
	}

	gasPrice, err := client.SuggestGasPrice(ctx)
	if err != nil {
		return "", err
	}

	tx := types.NewTransaction(nonce, toAddr, wei, 21000, gasPrice, nil)

	chainID, err := client.NetworkID(ctx)
	if err != nil {
		return "", err
	}
	signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privKey)
	if err != nil {
		return "", err
	}

	err = client.SendTransaction(ctx, signedTx)
	if err != nil {
		return "", err
	}

	return signedTx.Hash().Hex(), nil
}

func GetTransactionHistory(client *ethclient.Client, address string) ([]map[string]string, error) {
	apiKey := os.Getenv("ETHERSCAN_API_KEY")
	if apiKey == "" {
		return []map[string]string{}, fmt.Errorf("Etherscan API key not set")
	}

	url := fmt.Sprintf("https://api-sepolia.etherscan.io/api?module=account&action=txlist&address=%s&sort=desc&apikey=%s", address, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EtherscanResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if result.Status != "1" {
		return nil, fmt.Errorf("Etherscan API error: %s", result.Message)
	}

	transactions := make([]map[string]string, 0, len(result.Result))
	for _, tx := range result.Result {
		transactions = append(transactions, map[string]string{
			"hash":      tx.Hash,
			"from":      tx.From,
			"to":        tx.To,
			"value":     tx.Value,
			"gasUsed":   tx.GasUsed,
			"timestamp": tx.TimeStamp,
		})
	}

	return transactions, nil
}

func EstimateETHGas(client *ethclient.Client, fromAddress, toAddress, amount string) (string, error) {
	ctx := context.Background()
	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)

	amountFloat, ok := new(big.Float).SetString(amount)
	if !ok {
		return "", fmt.Errorf("invalid amount format")
	}
	wei := new(big.Int)
	amountFloat.Mul(amountFloat, big.NewFloat(1e18)).Int(wei)

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From:  fromAddr,
		To:    &toAddr,
		Value: wei,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", gasLimit), nil
}