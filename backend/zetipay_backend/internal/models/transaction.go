package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Transaction struct {
	ID            primitive.ObjectID `bson:"_id,omitempty"`
	WalletAddress string             `bson:"wallet_address"`
	TxHash        string             `bson:"tx_hash,unique"`
	Type          string             `bson:"type"`
	Amount        string             `bson:"amount"`
	ToAddress     string             `bson:"to_address"`
	TokenAddress  string             `bson:"token_address,omitempty"`
	CreatedAt     time.Time          `bson:"created_at"`
}