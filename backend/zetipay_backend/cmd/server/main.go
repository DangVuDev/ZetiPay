package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dangvu/zetipay/internal/config"
	// "github.com/gin-gonic/gin"
	// "crystapay/internal/config"
	// "crystapay/internal/handlers"
	// "crystapay/internal/middleware"
	// "crystapay/internal/services"
	// "crystapay/pkg/database"
)

func main() {
    // Load configuration
    cfg := config.Load()
    
    // Initialize database
    db, err := database.New(cfg.Database)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    
    // Initialize services
    blockchainService := services.NewBlockchainService(cfg.Blockchain)
    walletService := services.NewWalletService(db)
    transactionService := services.NewTransactionService(db, blockchainService)
    
    // Initialize handlers
    h := handlers.New(walletService, transactionService, blockchainService)
    
    // Setup router
    router := setupRouter(h)
    
    // Start server
    srv := &http.Server{
        Addr:    ":" + cfg.Server.Port,
        Handler: router,
    }
    
    // Graceful shutdown
    go func() {
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed to start: %v", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatal("Server forced to shutdown:", err)
    }
}

func setupRouter(h *handlers.Handler) *gin.Engine {
    r := gin.Default()
    
    // Middleware
    r.Use(middleware.CORS())
    r.Use(middleware.Logger())
    r.Use(middleware.ErrorHandler())
    
    // API routes
    api := r.Group("/api/v1")
    {
        // Wallet routes
        wallet := api.Group("/wallet")
        {
            wallet.POST("/connect", h.ConnectWallet)
            wallet.GET("/balance/:address", h.GetBalance)
            wallet.GET("/tokens/:address", h.GetTokens)
        }
        
        // Transaction routes
        tx := api.Group("/transactions")
        {
            tx.POST("/send", h.SendTransaction)
            tx.GET("/history/:address", h.GetTransactionHistory)
            tx.GET("/:hash", h.GetTransactionByHash)
            tx.GET("/pending/:address", h.GetPendingTransactions)
        }
        
        // Network routes
        network := api.Group("/network")
        {
            network.GET("/status", h.GetNetworkStatus)
            network.GET("/gas-price", h.GetGasPrice)
            network.GET("/supported", h.GetSupportedNetworks)
        }
    }
    
    return r
}