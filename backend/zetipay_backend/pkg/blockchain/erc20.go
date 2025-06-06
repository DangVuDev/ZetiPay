package blockchain

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

const erc20ABI = `[
	{"constant":true,"inputs":[{"name":"_owner","type":"address"}],"name":"balanceOf","outputs":[{"name":"balance","type":"uint256"}],"payable":false,"stateMutability":"view","type":"function"},
	{"constant":true,"inputs":[],"name":"decimals","outputs":[{"name":"","type":"uint8"}],"payable":false,"stateMutability":"view","type":"function"}
]`

func GetTokenBalance(client *ethclient.Client, address, tokenAddress string) (string, error) {
	ctx := context.Background()
	addr := common.HexToAddress(address)
	tokenAddr := common.HexToAddress(tokenAddress)

	parsedABI, err := abi.JSON(strings.NewReader(erc20ABI))
	if err != nil {
		return "", fmt.Errorf("failed to parse ERC20 ABI: %v", err)
	}

	// Get decimals
	dataDecimals, err := parsedABI.Pack("decimals")
	if err != nil {
		return "", fmt.Errorf("failed to pack decimals call: %v", err)
	}
	resultDecimals, err := client.CallContract(ctx, ethereum.CallMsg{To: &tokenAddr, Data: dataDecimals}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call decimals: %v", err)
	}
	var decimals uint8
	if err := parsedABI.UnpackIntoInterface(&decimals, "decimals", resultDecimals); err != nil {
		return "", fmt.Errorf("failed to unpack decimals: %v", err)
	}

	// Get balance
	dataBalance, err := parsedABI.Pack("balanceOf", addr)
	if err != nil {
		return "", fmt.Errorf("failed to pack balanceOf call: %v", err)
	}
	resultBalance, err := client.CallContract(ctx, ethereum.CallMsg{To: &tokenAddr, Data: dataBalance}, nil)
	if err != nil {
		return "", fmt.Errorf("failed to call balanceOf: %v", err)
	}
	var balance *big.Int
	if err := parsedABI.UnpackIntoInterface(&balance, "balanceOf", resultBalance); err != nil {
		return "", fmt.Errorf("failed to unpack balance: %v", err)
	}

	// Convert balance using decimals
	denominator := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	balanceFloat := new(big.Float).Quo(new(big.Float).SetInt(balance), new(big.Float).SetInt(denominator))
	return balanceFloat.Text('f', 6), nil
}

func SendTokenTransaction(client *ethclient.Client, fromAddress, toAddress, tokenAddress, amount, privateKey string) (string, error) {
	ctx := context.Background()
	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)
	tokenAddr := common.HexToAddress(tokenAddress)

	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount format")
	}

	// Placeholder for signature-based flow
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

	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	paddedAddress := common.LeftPadBytes(toAddr.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amountBig.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddr,
		To:   &tokenAddr,
		Data: data,
	})
	if err != nil {
		return "", fmt.Errorf("failed to estimate gas: %v", err)
	}

	tx := types.NewTransaction(nonce, tokenAddr, big.NewInt(0), gasLimit, gasPrice, data)

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

func EstimateTokenGas(client *ethclient.Client, fromAddress, toAddress, tokenAddress, amount string) (string, error) {
	ctx := context.Background()
	fromAddr := common.HexToAddress(fromAddress)
	toAddr := common.HexToAddress(toAddress)
	tokenAddr := common.HexToAddress(tokenAddress)

	amountBig, ok := new(big.Int).SetString(amount, 10)
	if !ok {
		return "", fmt.Errorf("invalid amount format")
	}

	transferFnSignature := []byte("transfer(address,uint256)")
	methodID := crypto.Keccak256(transferFnSignature)[:4]
	paddedAddress := common.LeftPadBytes(toAddr.Bytes(), 32)
	paddedAmount := common.LeftPadBytes(amountBig.Bytes(), 32)

	var data []byte
	data = append(data, methodID...)
	data = append(data, paddedAddress...)
	data = append(data, paddedAmount...)

	gasLimit, err := client.EstimateGas(ctx, ethereum.CallMsg{
		From: fromAddr,
		To:   &tokenAddr,
		Data: data,
	})
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d", gasLimit), nil
}