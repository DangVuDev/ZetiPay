package repository

import (
	"context"
	"time"

	"github.com/dangvu/zetipay/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func SaveTransaction(db *mongo.Database, tx *models.Transaction) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("transactions")

	// Create unique index on tx_hash
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "tx_hash", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	tx.CreatedAt = time.Now()
	_, err = collection.InsertOne(ctx, tx)
	return err
}

func GetTransactionsByAddress(db *mongo.Database, address string) ([]models.Transaction, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("transactions")
	cursor, err := collection.Find(ctx, bson.M{"wallet_address": address})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}