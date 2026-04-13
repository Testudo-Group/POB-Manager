package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrTransportNotFound = errors.New("transport not found")

type TransportRepository struct {
	collection *mongo.Collection
}

func NewTransportRepository(db *mongo.Database) *TransportRepository {
	return &TransportRepository{
		collection: db.Collection("transports"),
	}
}

func (r *TransportRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true).SetName("transports_name_unique"),
	}
	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *TransportRepository) Create(ctx context.Context, transport *domain.Transport) error {
	_, err := r.collection.InsertOne(ctx, transport)
	return err
}

func (r *TransportRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Transport, error) {
	var transport domain.Transport
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&transport)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrTransportNotFound
		}
		return nil, err
	}
	return &transport, nil
}

func (r *TransportRepository) FindAll(ctx context.Context, isActive *bool) ([]domain.Transport, error) {
	filter := bson.M{}
	if isActive != nil {
		filter["is_active"] = *isActive
	}
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transports []domain.Transport
	if err := cursor.All(ctx, &transports); err != nil {
		return nil, err
	}
	return transports, nil
}

func (r *TransportRepository) Update(ctx context.Context, transport *domain.Transport) error {
	filter := bson.M{"_id": transport.ID}
	update := bson.M{"$set": transport}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrTransportNotFound
	}
	return nil
}

func (r *TransportRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrTransportNotFound
	}
	return nil
}