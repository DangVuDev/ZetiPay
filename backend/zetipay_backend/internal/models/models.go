package models

import (
    "time"
    //"gorm.io/gorm"
)

type Wallet struct {
    ID        uint      `json:"id" gorm:"primaryKey"`
    Address   string    `json:"address" gorm:"uniqueIndex"`
    Network   string    `json:"network"`
    Balance   string    `json:"balance"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

type Transaction struct {
    ID          uint      `json:"id" gorm:"primaryKey"`
    Hash        string    `json:"hash" gorm:"uniqueIndex"`
    FromAddress string    `json:"from_address"`
    ToAddress   string    `json:"to_address"`
    Amount      string    `json:"amount"`
    Token       string    `json:"token"`
    Network     string    `json:"network"`
    Status      string    `json:"status"`
    GasUsed     string    `json:"gas_used"`
    GasPrice    string    `json:"gas_price"`
    BlockNumber uint64    `json:"block_number"`
    Timestamp   time.Time `json:"timestamp"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

type Token struct {
    ID           uint   `json:"id" gorm:"primaryKey"`
    Address      string `json:"address"`
    Symbol       string `json:"symbol"`
    Name         string `json:"name"`
    Decimals     int    `json:"decimals"`
    Network      string `json:"network"`
    LogoURL      string `json:"logo_url"`
    IsVerified   bool   `json:"is_verified"`
    PriceUSD     string `json:"price_usd"`
    MarketCap    string `json:"market_cap"`
    Volume24h    string `json:"volume_24h"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}

type NetworkStatus struct {
    ID              uint      `json:"id" gorm:"primaryKey"`
    Network         string    `json:"network"`
    IsActive        bool      `json:"is_active"`
    BlockNumber     uint64    `json:"block_number"`
    GasPrice        string    `json:"gas_price"`
    LastBlockTime   time.Time `json:"last_block_time"`
    PendingTxCount  int       `json:"pending_tx_count"`
    CreatedAt       time.Time `json:"created_at"`
    UpdatedAt       time.Time `json:"updated_at"`
}

// Response models
type SendTransactionRequest struct {
    FromAddress string `json:"from_address" binding:"required"`
    ToAddress   string `json:"to_address" binding:"required"`
    Amount      string `json:"amount" binding:"required"`
    Token       string `json:"token"`
    Network     string `json:"network" binding:"required"`
    GasLimit    string `json:"gas_limit"`
    GasPrice    string `json:"gas_price"`
}

type SendTransactionResponse struct {
    Hash        string `json:"hash"`
    Status      string `json:"status"`
    Message     string `json:"message"`
}

type BalanceResponse struct {
    Address string            `json:"address"`
    Network string            `json:"network"`
    Native  BalanceInfo       `json:"native"`
    Tokens  []TokenBalance    `json:"tokens"`
}

type BalanceInfo struct {
    Symbol   string `json:"symbol"`
    Balance  string `json:"balance"`
    USD      string `json:"usd"`
}

type TokenBalance struct {
    Address  string `json:"address"`
    Symbol   string `json:"symbol"`
    Balance  string `json:"balance"`
    Decimals int    `json:"decimals"`
    USD      string `json:"usd"`
}