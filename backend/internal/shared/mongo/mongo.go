package mongo

import (
	"context"

	"github.com/skyflow/skyflow/internal/shared/config"
	"github.com/skyflow/skyflow/internal/shared/logger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// NewClient creates a MongoDB client and returns the database
func NewClient(ctx context.Context, log *logger.Logger) (*mongo.Database, error) {
	uri := config.DB().Mongo
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	db := client.Database("skyflow")
	log.Info("mongodb connected", "uri", uri)
	return db, nil
}
