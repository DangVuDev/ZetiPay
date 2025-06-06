package repository

import (
	"context"
	"time"

	"github.com/dangvu/zetipay/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func UpsertUser(db *mongo.Database, user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := db.Collection("users")

	// Create unique index on wallet_address
	_, err := collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "wallet_address", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}

	user.UpdatedAt = time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now()
	}

	_, err = collection.UpdateOne(
		ctx,
		bson.M{"wallet_address": user.WalletAddress},
		bson.M{
			"$set": bson.M{
				"wallet_address": user.WalletAddress,
				"updated_at":     user.UpdatedAt,
				"created_at":     user.CreatedAt,
			},
		},
		options.Update().SetUpsert(true),
	)
	return err
}