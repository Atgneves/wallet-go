package database

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoClient struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func NewMongoClient(uri string) (*MongoClient, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Test the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	database := client.Database("wallet")

	return &MongoClient{
		Client:   client,
		Database: database,
	}, nil
}

func (mc *MongoClient) Disconnect(ctx context.Context) error {
	return mc.Client.Disconnect(ctx)
}

func (mc *MongoClient) GetCollection(name string) *mongo.Collection {
	return mc.Database.Collection(name)
}
