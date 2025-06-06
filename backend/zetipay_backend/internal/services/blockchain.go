package services

import (
	"github.com/dangvu/zetipay/pkg/blockchain"
	"github.com/ethereum/go-ethereum/ethclient"
)

type BlockchainService struct {
	client *ethclient.Client
}

func NewBlockchainService(providerURL string) *BlockchainService {
	client, err := ethclient.Dial(providerURL)
	if err != nil {
		panic(err) // Handle properly in production
	}
	return &BlockchainService{client: client}
}

func (s *BlockchainService) GetBalances(address string) (string, map[string]string, error) {
	return blockchain.GetBalances(s.client, address)
}

func (s *BlockchainService) SendETHTransaction(fromAddress, toAddress, amount, privateKey string) (string, error) {
	return blockchain.SendETHTransaction(s.client, fromAddress, toAddress, amount, privateKey)
}

func (s *BlockchainService) SendTokenTransaction(fromAddress, toAddress, tokenAddress, amount, privateKey string) (string, error) {
	return blockchain.SendTokenTransaction(s.client, fromAddress, toAddress, tokenAddress, amount, privateKey)
}

func (s *BlockchainService) GetTransactionHistory(address string) ([]map[string]string, error) {
	return blockchain.GetTransactionHistory(s.client, address)
}

func (s *BlockchainService) EstimateETHGas(fromAddress, toAddress, amount string) (string, error) {
	return blockchain.EstimateETHGas(s.client, fromAddress, toAddress, amount)
}

func (s *BlockchainService) EstimateTokenGas(fromAddress, toAddress, tokenAddress, amount string) (string, error) {
	return blockchain.EstimateTokenGas(s.client, fromAddress, toAddress, tokenAddress, amount)
}