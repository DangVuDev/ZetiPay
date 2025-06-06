package services

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"strconv"

	"github.com/dangvu/zetipay/internal/config"
	"github.com/dangvu/zetipay/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainService struct {
    clients map[string]*ethclient.Client
    configs map[string]interface{}
}

func NewBlockchainService(config config.BlockchainConfig) *BlockchainService {
    clients := make(map[string]*ethclient.Client)
    configs := make(map[string]interface{})
    
    // Initialize Ethereum client
    if config.Ethereum.RPCURL != "" {
        ethClient, err := ethclient.Dial(config.Ethereum.RPCURL)
        if err == nil {
            clients["ethereum"] = ethClient
            configs["ethereum"] = config.Ethereum
        }
    }
    
    // Initialize BSC client
    if config.BSC.RPCURL != "" {
        bscClient, err := ethclient.Dial(config.BSC.RPCURL)
        if err == nil {
            clients["bsc"] = bscClient
            configs["bsc"] = config.BSC
        }
    }
    
    // Initialize Polygon client
    if config.Polygon.RPCURL != "" {
        polygonClient, err := ethclient.Dial(config.Polygon.RPCURL)
        if err == nil {
            clients["polygon"] = polygonClient
            configs["polygon"] = config.Polygon
        }
    }
    
    return &BlockchainService{
        clients: clients,
        configs: configs,
    }
}

func (s *BlockchainService) GetBalance(network, address string) (*big.Int, error) {
    client, exists := s.clients[network]
    if !exists {
        return nil, fmt.Errorf("network %s not supported", network)
    }
    
    account := common.HexToAddress(address)
    balance, err := client.BalanceAt(context.Background(), account, nil)
    if err != nil {
        return nil, err
    }
    
    return balance, nil
}

func (s *BlockchainService) SendTransaction(req *models.SendTransactionRequest) (*models.SendTransactionResponse, error) {
    client, exists := s.clients[req.Network]
    if !exists {
        return nil, fmt.Errorf("network %s not supported", req.Network)
    }
    
    // Get private key from config
    var privateKeyHex string
    switch req.Network {
    case "ethereum":
        cfg := s.configs["ethereum"].(config.EthereumConfig)
        privateKeyHex = cfg.PrivateKey
    case "bsc":
        cfg := s.configs["bsc"].(config.BSCConfig)
        privateKeyHex = cfg.PrivateKey
    case "polygon":
        cfg := s.configs["polygon"].(config.PolygonConfig)
        privateKeyHex = cfg.PrivateKey
    default:
        return nil, fmt.Errorf("unsupported network: %s", req.Network)
    }
    
    privateKey, err := crypto.HexToECDSA(privateKeyHex)
    if err != nil {
        return nil, err
    }
    
    publicKey := privateKey.Public().(*ecdsa.PublicKey)
    fromAddress := crypto.PubkeyToAddress(*publicKey)
    
    nonce, err := client.PendingNonceAt(context.Background(), fromAddress)
    if err != nil {
        return nil, err
    }
    
    value := new(big.Int)
    value.SetString(req.Amount, 10)
    
    gasLimit := uint64(21000) // Standard gas limit for simple transfer
    if req.GasLimit != "" {
        gasLimit, _ = strconv.ParseUint(req.GasLimit, 10, 64)
    }
    
    gasPrice, err := client.SuggestGasPrice(context.Background())
    if err != nil {
        return nil, err
    }
    
    if req.GasPrice != "" {
        gasPrice = new(big.Int)
        gasPrice.SetString(req.GasPrice, 10)
    }
    
    toAddress := common.HexToAddress(req.ToAddress)
    
    tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, nil)
    
    chainID := big.NewInt(1) // Default to Ethereum mainnet
    switch req.Network {
    case "bsc":
        chainID = big.NewInt(56)
    case "polygon":
        chainID = big.NewInt(137)
    }
    
    signedTx, err := types.SignTx(tx, types.NewEIP155Signer(chainID), privateKey)
    if err != nil {
        return nil, err
    }
    
    err = client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return nil, err
    }
    
    return &models.SendTransactionResponse{
        Hash:    signedTx.Hash().Hex(),
        Status:  "pending",
        Message: "Transaction sent successfully",
    }, nil
}

func (s *BlockchainService) GetTransactionByHash(network, hash string) (*types.Transaction, error) {
    client, exists := s.clients[network]
    if !exists {
        return nil, fmt.Errorf("network %s not supported", network)
    }
    
    txHash := common.HexToHash(hash)
    tx, _, err := client.TransactionByHash(context.Background(), txHash)
    if err != nil {
        return nil, err
    }
    
    return tx, nil
}

func (s *BlockchainService) GetGasPrice(network string) (*big.Int, error) {
    client, exists := s.clients[network]
    if !exists {
        return nil, fmt.Errorf("network %s not supported", network)
    }
    
    return client.SuggestGasPrice(context.Background())
}

func (s *BlockchainService) GetBlockNumber(network string) (uint64, error) {
    client, exists := s.clients[network]
    if !exists {
        return 0, fmt.Errorf("network %s not supported", network)
    }
    
    return client.BlockNumber(context.Background())
}