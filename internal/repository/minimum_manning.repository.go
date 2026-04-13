package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrMinimumManningEventNotFound = errors.New("minimum manning event not found")

type MinimumManningRepository struct {
	collection *mongo.Collection
}

func NewMinimumManningRepository(db *mongo.Database) *MinimumManningRepository {
	return &MinimumManningRepository{
		collection: db.Collection("minimum_manning_events"),
	}
}

func (r *MinimumManningRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "vessel_id", Value: 1}, {Key: "status", Value: 1}}},
		{Keys: bson.D{{Key: "activated_at", Value: -1}}},
	}
	for _, idx := range indexes {
		_, err := r.collection.Indexes().CreateOne(ctx, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *MinimumManningRepository) Create(ctx context.Context, event *domain.MinimumManningEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *MinimumManningRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.MinimumManningEvent, error) {
	var event domain.MinimumManningEvent
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&event)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrMinimumManningEventNotFound
		}
		return nil, err
	}
	return &event, nil
}

func (r *MinimumManningRepository) FindActiveByVessel(ctx context.Context, vesselID bson.ObjectID) (*domain.MinimumManningEvent, error) {
	var event domain.MinimumManningEvent
	filter := bson.M{
		"vessel_id": vesselID,
		"status":    domain.MinimumManningStatusActive,
	}
	err := r.collection.FindOne(ctx, filter).Decode(&event)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // No active event
		}
		return nil, err
	}
	return &event, nil
}

func (r *MinimumManningRepository) FindByVessel(ctx context.Context, vesselID bson.ObjectID, limit int64) ([]domain.MinimumManningEvent, error) {
	filter := bson.M{"vessel_id": vesselID}
	opts := options.Find().SetSort(bson.D{{Key: "activated_at", Value: -1}}).SetLimit(limit)
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []domain.MinimumManningEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}

func (r *MinimumManningRepository) Update(ctx context.Context, event *domain.MinimumManningEvent) error {
	filter := bson.M{"_id": event.ID}
	update := bson.M{"$set": event}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrMinimumManningEventNotFound
	}
	return nil
}