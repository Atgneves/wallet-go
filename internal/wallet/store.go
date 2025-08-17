package wallet

import (
	"context"
	"fmt"
	"time"

	"wallet-go/internal/operation"
	"wallet-go/internal/shared/database"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	collection          *mongo.Collection
	operationCollection *mongo.Collection
}

func NewStore(db *database.MongoClient) *Store {
	return &Store{
		collection:          db.GetCollection("wallet"),
		operationCollection: db.GetCollection("operation"),
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
	filter := bson.M{"walletId": walletID}

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

	// Carregar operações da carteira
	if err := s.loadOperations(ctx, &wallet); err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (s *Store) FindAll(ctx context.Context) ([]*Wallet, error) {
	fmt.Printf("DEBUG: Store.FindAll() - iniciando busca no MongoDB...\n")

	cursor, err := s.collection.Find(ctx, bson.M{})
	if err != nil {
		fmt.Printf("DEBUG: Store.FindAll() - erro na query MongoDB: %v\n", err)
		return nil, err
	}
	defer cursor.Close(ctx)

	fmt.Printf("DEBUG: Store.FindAll() - cursor criado com sucesso, iterando...\n")

	var wallets []*Wallet
	for cursor.Next(ctx) {
		var wallet Wallet
		if err := cursor.Decode(&wallet); err != nil {
			fmt.Printf("DEBUG: Store.FindAll() - erro ao decodificar wallet: %v\n", err)
			return nil, err
		}

		fmt.Printf("DEBUG: Store.FindAll() - wallet decodificada: %s\n", wallet.WalletID)
		wallets = append(wallets, &wallet)
	}

	fmt.Printf("DEBUG: Store.FindAll() - finalizando, total de wallets: %d\n", len(wallets))
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

// loadOperations carrega as operações da carteira ordenadas por data de criação
func (s *Store) loadOperations(ctx context.Context, wallet *Wallet) error {
	filter := bson.M{"walletId": wallet.WalletID}

	// Ordenar por data de criação (mais recente primeiro)
	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := s.operationCollection.Find(ctx, filter, opts)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	var operations []operation.Operation
	for cursor.Next(ctx) {
		var op operation.Operation
		if err := cursor.Decode(&op); err != nil {
			return err
		}
		operations = append(operations, op)
	}

	wallet.Operations = operations
	return cursor.Err()
}

// FindByIDWithoutOperations busca carteira sem carregar operações (para performance)
func (s *Store) FindByIDWithoutOperations(ctx context.Context, walletID uuid.UUID) (*Wallet, error) {
	var wallet Wallet
	filter := bson.M{"walletId": walletID}

	err := s.collection.FindOne(ctx, filter).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &wallet, nil
}
