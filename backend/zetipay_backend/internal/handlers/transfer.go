package handlers

import (
	"net/http"

	"github.com/dangvu/zetipay/internal/models"
	"github.com/dangvu/zetipay/internal/repository"
	"github.com/dangvu/zetipay/pkg/blockchain"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// TransferETHRequest defines the request body for ETH transfer
type TransferETHRequest struct {
	FromAddress string `json:"fromAddress" binding:"required"`
	ToAddress   string `json:"toAddress" binding:"required"`
	Amount      string `json:"amount" binding:"required"`
	Signature   string `json:"signature" binding:"required"`
}

// TransferTokenRequest defines the request body for token transfer
type TransferTokenRequest struct {
	FromAddress  string `json:"fromAddress" binding:"required"`
	ToAddress    string `json:"toAddress" binding:"required"`
	TokenAddress string `json:"tokenAddress" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	Signature    string `json:"signature" binding:"required"`
}

// TransferResponse defines the response body for transfers
type TransferResponse struct {
	TransactionHash string `json:"transactionHash"`
}

// TransferETH godoc
// @Summary Transfer ETH
// @Description Initiates an Ethereum transfer
// @Tags Transfer
// @Accept json
// @Produce json
// @Param request body TransferETHRequest true "ETH transfer details"
// @Success 200 {object} TransferResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security Bearer
// @Router /transfer/eth [post]
func TransferETH(client *ethclient.Client, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TransferETHRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !common.IsHexAddress(req.FromAddress) || !common.IsHexAddress(req.ToAddress) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
			return
		}

		// TODO: Verify EIP-712 signature
		txHash, err := blockchain.SendETHTransaction(client, req.FromAddress, req.ToAddress, req.Amount, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Save transaction metadata
		tx := models.Transaction{
			WalletAddress: req.FromAddress,
			TxHash:        txHash,
			Type:          "ETH_TRANSFER",
			Amount:        req.Amount,
			ToAddress:     req.ToAddress,
		}
		if err := repository.SaveTransaction(db, &tx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save transaction"})
			return
		}

		c.JSON(http.StatusOK, TransferResponse{TransactionHash: txHash})
	}
}

// TransferToken godoc
// @Summary Transfer ERC-20 token
// @Description Initiates an ERC-20 token transfer
// @Tags Transfer
// @Accept json
// @Produce json
// @Param request body TransferTokenRequest true "Token transfer details"
// @Success 200 {object} TransferResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Security Bearer
// @Router /transfer/token [post]
func TransferToken(client *ethclient.Client, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req TransferTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !common.IsHexAddress(req.FromAddress) || !common.IsHexAddress(req.ToAddress) || !common.IsHexAddress(req.TokenAddress) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
			return
		}

		// TODO: Verify EIP-712 signature
		txHash, err := blockchain.SendTokenTransaction(client, req.FromAddress, req.ToAddress, req.TokenAddress, req.Amount, "")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Save transaction metadata
		tx := models.Transaction{
			WalletAddress: req.FromAddress,
			TxHash:        txHash,
			Type:          "TOKEN_TRANSFER",
			Amount:        req.Amount,
			ToAddress:     req.ToAddress,
			TokenAddress:  req.TokenAddress,
		}
		if err := repository.SaveTransaction(db, &tx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save transaction"})
			return
		}

		c.JSON(http.StatusOK, TransferResponse{TransactionHash: txHash})
	}
}