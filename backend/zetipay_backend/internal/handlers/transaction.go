package handlers

import (
	"log"
	"math/big"
	"net/http"
	"strconv"

	"github.com/dangvu/zetipay/internal/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/gin-gonic/gin"
)

type Handler struct {
    walletService      *services.WalletService
    transactionService *services.TransactionService
    blockchainService  *services.BlockchainService
}

func New(ws *services.WalletService, ts *services.TransactionService, bs *services.BlockchainService) *Handler {
    return &Handler{
        walletService:      ws,
        transactionService: ts,
        blockchainService:  bs,
    }
}

func (h *Handler) SendTransaction(c *gin.Context) {
    var req models.SendTransactionRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format",
            "details": err.Error(),
        })
        return
    }
    
    // Validate addresses
    if !isValidAddress(req.FromAddress) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid from address",
        })
        return
    }
    
    if !isValidAddress(req.ToAddress) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid to address",
        })
        return
    }
    
    // Send transaction through blockchain service
    response, err := h.blockchainService.SendTransaction(&req)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to send transaction",
            "details": err.Error(),
        })
        return
    }
    
    // Save transaction to database
    tx := &models.Transaction{
        Hash:        response.Hash,
        FromAddress: req.FromAddress,
        ToAddress:   req.ToAddress,
        Amount:      req.Amount,
        Token:       req.Token,
        Network:     req.Network,
        Status:      "pending",
        GasPrice:    req.GasPrice,
    }
    
    if err := h.transactionService.SaveTransaction(tx); err != nil {
        // Log error but don't fail the request since transaction was already sent
        log.Printf("Failed to save transaction to database: %v", err)
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *Handler) GetTransactionHistory(c *gin.Context) {
    address := c.Param("address")
    if !isValidAddress(address) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid address",
        })
        return
    }
    
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
    network := c.Query("network")
    
    transactions, total, err := h.transactionService.GetTransactionHistory(address, network, page, limit)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get transaction history",
            "details": err.Error(),
        })
        return
    }
    
    c.JSON(http.StatusOK, gin.H{
        "transactions": transactions,
        "total": total,
        "page": page,
        "limit": limit,
    })
}

func (h *Handler) GetTransactionByHash(c *gin.Context) {
    hash := c.Param("hash")
    network := c.Query("network")
    
    if network == "" {
        network = "ethereum" // default
    }
    
    transaction, err := h.transactionService.GetTransactionByHash(hash)
    if err != nil {
        // Try to get from blockchain if not in database
        blockchainTx, err := h.blockchainService.GetTransactionByHash(network, hash)
        if err != nil {
            c.JSON(http.StatusNotFound, gin.H{
                "error": "Transaction not found",
            })
            return
        }
        
        // Convert blockchain transaction to our model
        // This is a simplified conversion
        c.JSON(http.StatusOK, gin.H{
            "hash": blockchainTx.Hash().Hex(),
            "status": "confirmed",
            "gas_price": blockchainTx.GasPrice().String(),
            "gas_limit": strconv.FormatUint(blockchainTx.Gas(), 10),
        })
        return
    }
    
    c.JSON(http.StatusOK, transaction)
}

func (h *Handler) GetBalance(c *gin.Context) {
    address := c.Param("address")
    network := c.DefaultQuery("network", "ethereum")
    
    if !isValidAddress(address) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid address",
        })
        return
    }
    
    balance, err := h.blockchainService.GetBalance(network, address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get balance",
            "details": err.Error(),
        })
        return
    }
    
    // Convert wei to ether for display
    ethBalance := new(big.Float)
    ethBalance.SetString(balance.String())
    ethBalance.Quo(ethBalance, big.NewFloat(1e18))
    
    response := &models.BalanceResponse{
        Address: address,
        Network: network,
        Native: models.BalanceInfo{
            Symbol:  getNetworkSymbol(network),
            Balance: ethBalance.String(),
            USD:     "0", // TODO: Get price from external API
        },
        Tokens: []models.TokenBalance{}, // TODO: Get token balances
    }
    
    c.JSON(http.StatusOK, response)
}

func (h *Handler) GetNetworkStatus(c *gin.Context) {
    networks := []string{"ethereum", "bsc", "polygon"}
    status := make(map[string]interface{})
    
    for _, network := range networks {
        blockNumber, err := h.blockchainService.GetBlockNumber(network)
        if err != nil {
            status[network] = map[string]interface{}{
                "is_active": false,
                "error": err.Error(),
            }
            continue
        }
        
        gasPrice, err := h.blockchainService.GetGasPrice(network)
        if err != nil {
            gasPrice = big.NewInt(0)
        }
        
        status[network] = map[string]interface{}{
            "is_active": true,
            "block_number": blockNumber,
            "gas_price": gasPrice.String(),
        }
    }
    
    c.JSON(http.StatusOK, status)
}

func (h *Handler) GetGasPrice(c *gin.Context) {
    network := c.DefaultQuery("network", "ethereum")
    
    gasPrice, err := h.blockchainService.GetGasPrice(network)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to get gas price",
            "details": err.Error(),
        })
        return
    }
    
    // Convert to Gwei
    gwei := new(big.Float)
    gwei.SetString(gasPrice.String())
    gwei.Quo(gwei, big.NewFloat(1e9))
    
    c.JSON(http.StatusOK, gin.H{
        "network": network,
        "gas_price_wei": gasPrice.String(),
        "gas_price_gwei": gwei.String(),
    })
}

func (h *Handler) ConnectWallet(c *gin.Context) {
    var req struct {
        Address string `json:"address" binding:"required"`
        Network string `json:"network" binding:"required"`
    }
    
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid request format",
        })
        return
    }
    
    if !isValidAddress(req.Address) {
        c.JSON(http.StatusBadRequest, gin.H{
            "error": "Invalid address",
        })
        return
    }
    
    // Get balance for the wallet
    balance, err := h.blockchainService.GetBalance(req.Network, req.Address)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": "Failed to connect wallet",
            "details": err.Error(),
        })
        return
    }
    
    // Save or update wallet in database
    wallet := &models.Wallet{
        Address: req.Address,
        Network: req.Network,
        Balance: balance.String(),
    }
    
    if err := h.walletService.SaveWallet(wallet); err != nil {
        log.Printf("Failed to save wallet: %v", err)
    }
    
    c.JSON(http.StatusOK, gin.H{
        "message": "Wallet connected successfully",
        "address": req.Address,
        "network": req.Network,
        "balance": balance.String(),
    })
}

func (h *Handler) GetSupportedNetworks(c *gin.Context) {
    networks := []map[string]interface{}{
        {
            "id": "ethereum",
            "name": "Ethereum",
            "symbol": "ETH",
            "chain_id": 1,
            "rpc_url": "https://mainnet.infura.io/v3/",
            "explorer": "https://etherscan.io",
        },
        {
            "id": "bsc",
            "name": "Binance Smart Chain",
            "symbol": "BNB",
            "chain_id": 56,
            "rpc_url": "https://bsc-dataseed.binance.org/",
            "explorer": "https://bscscan.com",
        },
        {
            "id": "polygon",
            "name": "Polygon",
            "symbol": "MATIC",
            "chain_id": 137,
            "rpc_url": "https://polygon-rpc.com/",
            "explorer": "https://polygonscan.com",
        },
    }
    
    c.JSON(http.StatusOK, gin.H{
        "networks": networks,
    })
}

// Helper functions
func isValidAddress(address string) bool {
    return common.IsHexAddress(address)
}

func getNetworkSymbol(network string) string {
    symbols := map[string]string{
        "ethereum": "ETH",
        "bsc":      "BNB",
        "polygon":  "MATIC",
    }
    
    if symbol, exists := symbols[network]; exists {
        return symbol
    }
    return "ETH"
}