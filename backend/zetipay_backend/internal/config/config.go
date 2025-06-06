package config

import (
    "os"
    "strconv"
)

type Config struct {
    Server     ServerConfig
    Database   DatabaseConfig
    Blockchain BlockchainConfig
    Redis      RedisConfig
}

type ServerConfig struct {
    Port string
    Env  string
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    DBName   string
    SSLMode  string
}

type BlockchainConfig struct {
    Ethereum EthereumConfig
    BSC      BSCConfig
    Polygon  PolygonConfig
}

type EthereumConfig struct {
    RPCURL     string
    PrivateKey string
    ChainID    int
}

type BSCConfig struct {
    RPCURL     string
    PrivateKey string
    ChainID    int
}

type PolygonConfig struct {
    RPCURL     string
    PrivateKey string
    ChainID    int
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

func Load() *Config {
    dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "5432"))
    redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
    
    return &Config{
        Server: ServerConfig{
            Port: getEnv("PORT", "8080"),
            Env:  getEnv("ENV", "development"),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     dbPort,
            User:     getEnv("DB_USER", "postgres"),
            Password: getEnv("DB_PASSWORD", ""),
            DBName:   getEnv("DB_NAME", "crystapay"),
            SSLMode:  getEnv("DB_SSLMODE", "disable"),
        },
        Blockchain: BlockchainConfig{
            Ethereum: EthereumConfig{
                RPCURL:     getEnv("ETHEREUM_RPC_URL", ""),
                PrivateKey: getEnv("ETHEREUM_PRIVATE_KEY", ""),
                ChainID:    1,
            },
            BSC: BSCConfig{
                RPCURL:     getEnv("BSC_RPC_URL", ""),
                PrivateKey: getEnv("BSC_PRIVATE_KEY", ""),
                ChainID:    56,
            },
            Polygon: PolygonConfig{
                RPCURL:     getEnv("POLYGON_RPC_URL", ""),
                PrivateKey: getEnv("POLYGON_PRIVATE_KEY", ""),
                ChainID:    137,
            },
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       redisDB,
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}