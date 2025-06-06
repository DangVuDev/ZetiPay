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

// WalletConnectRequest defines the request body for wallet connection
type WalletConnectRequest struct {
	Address string `json:"address" binding:"required"`
}

// WalletConnectResponse defines the response body for wallet connection
type WalletConnectResponse struct {
	Address       string            `json:"address"`
	ETHBalance    string            `json:"ethBalance"`
	TokenBalances map[string]string `json:"tokenBalances"`
}

// WalletConnect godoc
// @Summary Connect a wallet
// @Description Connects an Ethereum wallet and retrieves its balances
// @Tags Wallet
// @Accept json
// @Produce json
// @Param request body WalletConnectRequest true "Wallet address"
// @Success 200 {object} WalletConnectResponse
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /wallet/connect [post]
func WalletConnect(client *ethclient.Client, db *mongo.Database) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req WalletConnectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
			return
		}

		if !common.IsHexAddress(req.Address) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid wallet address"})
			return
		}

		ethBalance, tokenBalances, err := blockchain.GetBalances(client, req.Address)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch balances"})
			return
		}

		// Save or update user in database
		user := models.User{WalletAddress: req.Address}
		if err := repository.UpsertUser(db, &user); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save user"})
			return
		}

		c.JSON(http.StatusOK, WalletConnectResponse{
			Address:       req.Address,
			ETHBalance:    ethBalance,
			TokenBalances: tokenBalances,
		})
	}
}