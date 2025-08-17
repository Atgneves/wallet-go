package operation

import (
	"context"
	"time"

	"wallet-go/internal/shared/database"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Store struct {
	collection *mongo.Collection
}

func NewStore(db *database.MongoClient) *Store {
	return &Store{
		collection: db.GetCollection("operation"),
	}
}

func (s *Store) Create(ctx context.Context, operation *Operation) error {
	operation.CreatedAt = time.Now()

	_, err := s.collection.InsertOne(ctx, operation)
	return err
}

func (s *Store) FindByWalletID(ctx context.Context, walletID uuid.UUID) ([]*Operation, error) {
	filter := bson.M{"walletId": walletID}

	cursor, err := s.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operations []*Operation
	for cursor.Next(ctx) {
		var operation Operation
		if err := cursor.Decode(&operation); err != nil {
			return nil, err
		}
		operations = append(operations, &operation)
	}

	return operations, cursor.Err()
}

func (s *Store) FindByID(ctx context.Context, operationID uuid.UUID) (*Operation, error) {
	var operation Operation
	filter := bson.M{"operationId": operationID}

	err := s.collection.FindOne(ctx, filter).Decode(&operation)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &operation, nil
}

func (s *Store) FindByWalletIDAndDateRange(ctx context.Context, walletID uuid.UUID, from, to time.Time) ([]*Operation, error) {
	filter := bson.M{
		"walletId": walletID,
		"createdAt": bson.M{
			"$gte": from,
			"$lte": to,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operations []*Operation
	for cursor.Next(ctx) {
		var operation Operation
		if err := cursor.Decode(&operation); err != nil {
			return nil, err
		}
		operations = append(operations, &operation)
	}

	return operations, cursor.Err()
}

func (s *Store) FindByWalletIDAndDate(ctx context.Context, walletID uuid.UUID, date time.Time) ([]*Operation, error) {
	startOfDay := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	filter := bson.M{
		"walletId": walletID,
		"createdAt": bson.M{
			"$gte": startOfDay,
			"$lt":  endOfDay,
		},
	}

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := s.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var operations []*Operation
	for cursor.Next(ctx) {
		var operation Operation
		if err := cursor.Decode(&operation); err != nil {
			return nil, err
		}
		operations = append(operations, &operation)
	}

	return operations, cursor.Err()
}
