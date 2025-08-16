package wallet

import (
	"context"
	"time"

	"wallet-go/internal/shared/database"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Store struct {
	collection *mongo.Collection
}

func NewStore(db *database.MongoClient) *Store {
	return &Store{
		collection: db.GetCollection("wallet"),
	}
}

func (s *Store) Create(ctx context.Context, wallet *Wallet) error {
	wallet.CreatedAt = time.Now()
	wallet.UpdatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, wallet)
	return err
}

func (s *Store) FindByID(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	var wallet Wallet
	filter := bson.M{"WalletID": walletID}

	err := s.collection.FindOne(ctx, filter).Decode(&wallet)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &wallet, nil
}

func (s *Store) FindByCustomerID(ctx context.Context, customerID string) (*Wallet, error) {
	var wallet Wallet
	filter := bson.M{"customerId": customerID}

	err := s.collection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &wallet, nil
}

func (s *Store) FindAll(ctx context.Context) ([]*Wallet, error) {
	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []*Wallet
	for cursor.Next(ctx) {
		var wallet Wallet
		if err := cursor.Decode(&wallet); err != nil {
			return nil, err
		}
		wallets = append(wallets, &wallet)
	}

	return wallets, cursor.Err()
}

func (s *Store) Update(ctx context.Context, wallet *Wallet) error {
	wallet.UpdatedAt = time.Now()

	filter := bson.M{"walletId": wallet.WalletID}
	update := bson.M{"$set": wallet}

	_, err := s.collection.UpdateOne(ctx, filter, update)
	return err
}

func (s *Store) Delete(ctx context.Context, walletID uuid.UUID) error {
	filter := bson.M{"walletId": walletID}

	_, err := s.collection.DeleteOne(ctx, filter)
	return err
}
