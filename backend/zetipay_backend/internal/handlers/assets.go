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

// GasEstimateRequest defines the request body for gas estimation
type GasEstimateRequest struct {
	FromAddress  string `json:"fromAddress" binding:"required"`
	ToAddress    string `json:"toAddress" binding:"required"`
	Amount       string `json:"amount" binding:"required"`
	TokenAddress string `json:"tokenAddress,omitempty"`
}

// AssetsResponse defines the response body for asset queries
type AssetsResponse struct {
	Address        string                `json:"address"`
	ETHBalance     string                `json:"ethBalance"`
	TokenBalances  map[string]string     `json:"tokenBalances"`
	Transactions   []map[string]string `json:"transactions"`
	DBTransactions []models.Transaction  `json:"dbTransactions"`
}

// GasEstimateResponse defines the response body for gas estimation
type GasEstimateResponse struct {
	GasEstimate string `json:"gasEstimate"`
}

// GetAssets godoc
// @Summary Get wallet assets
// @Description Retrieves ETH and token balances, transaction history
// @Tags Assets
// @Produce json
// @Param address path string true "Wallet address"
// @Success 200 {object} AssetsResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /assets/{address} [get]
func GetAssets(client *ethclient.Client, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		address := c.Param("address")
		if !common.IsHexAddress(address) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet address"})
			return
		}

		ethBalance, tokenBalances, err := blockchain.GetBalances(client, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balances"})
			return
		}

		transactions, err := blockchain.GetTransactionHistory(client, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch transaction history"})
			return
		}

		dbTransactions, err := repository.GetTransactionsByAddress(db, address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch database transactions"})
			return
		}

		c.JSON(http.StatusOK, AssetsResponse{
			Address:        address,
			ETHBalance:     ethBalance,
			TokenBalances:  tokenBalances,
			Transactions:   transactions,
			DBTransactions: dbTransactions,
		})
	}
}

// EstimateGas godoc
// @Summary Estimate gas for a transaction
// @Description Estimates gas for ETH or token transfer
// @Tags Assets
// @Accept json
// @Produce json
// @Param request body GasEstimateRequest true "Transaction details"
// @Success 200 {object} GasEstimateResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /gas/estimate [post]
func EstimateGas(client *ethclient.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req GasEstimateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !common.IsHexAddress(req.FromAddress) || !common.IsHexAddress(req.ToAddress) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address"})
			return
		}

		var gasEstimate string
		var err error
		if req.TokenAddress == "" {
			gasEstimate, err = blockchain.EstimateETHGas(client, req.FromAddress, req.ToAddress, req.Amount)
		} else {
			if !common.IsHexAddress(req.TokenAddress) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token address"})
				return
			}
			gasEstimate, err = blockchain.EstimateTokenGas(client, req.FromAddress, req.ToAddress, req.TokenAddress, req.Amount)
		}

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, GasEstimateResponse{GasEstimate: gasEstimate})
	}
}