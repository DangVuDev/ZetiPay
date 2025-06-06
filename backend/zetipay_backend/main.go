package main

import (
	"context"
	"log"
	"time"

	"github.com/dangvu/zetipay/internal/config"
	"github.com/dangvu/zetipay/internal/handlers"
	"github.com/dangvu/zetipay/internal/middleware"
	"github.com/dangvu/zetipay/pkg/database"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/swag/example/basic/docs"
)

// @title ZetiPay API
// @version 1.0
// @description ZetiPay REST API for wallet connection, crypto transfers, and asset management.
// @host localhost:5000
// @BasePath /api
func main() {
	// Load environment variables
	if err := godotenv.Load(".env"); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}

	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate Web3 provider URL
	if cfg.Web3ProviderURL == "" {
		log.Fatal("WEB3_PROVIDER_URL is not set in environment variables")
	}

	// Initialize Ethereum client
	client, err := ethclient.Dial(cfg.Web3ProviderURL)
	if err != nil {
		log.Fatalf("Failed to connect to Ethereum client: %v", err)
	}

	// Initialize MongoDB client
	if cfg.MongoURI == "" {
		log.Fatal("MONGO_URI is not set in environment variables")
	}
	mongoClient, err := database.NewMongoDB(cfg.MongoURI)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Printf("Failed to disconnect MongoDB: %v", err)
		}
	}()

	// Get database and collections
	db := mongoClient.Database("zetipay")

	// Initialize Gin router
	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())
	r.Use(middleware.CORS())

	// Swagger route
	docs.SwaggerInfo.BasePath = "/api"
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes
	api := r.Group("/api")
	{
		api.POST("/wallet/connect", handlers.WalletConnect(client, db))
		api.POST("/transfer/eth", middleware.Auth(), handlers.TransferETH(client, db))
		api.POST("/transfer/token", middleware.Auth(), handlers.TransferToken(client, db))
		api.GET("/assets/:address", handlers.GetAssets(client, db))
		api.POST("/gas/estimate", handlers.EstimateGas(client))
	}

	// Start server
	log.Printf("REST API running on port %s", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}