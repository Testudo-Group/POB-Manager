package repository

import (
	"context"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type VesselEventRepository struct {
	collection *mongo.Collection
}

func NewVesselEventRepository(db *mongo.Database) *VesselEventRepository {
	return &VesselEventRepository{
		collection: db.Collection("vessel_events"),
	}
}

func (r *VesselEventRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "vessel_id", Value: 1},
			{Key: "occurred_at", Value: -1},
		},
	})
	return err
}

func (r *VesselEventRepository) Create(ctx context.Context, event *domain.VesselEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

func (r *VesselEventRepository) FindByVessel(ctx context.Context, vesselID bson.ObjectID, limit int64) ([]domain.VesselEvent, error) {
	opts := options.Find().SetSort(bson.D{{Key: "occurred_at", Value: -1}})
	if limit > 0 {
		opts.SetLimit(limit)
	}

	cursor, err := r.collection.Find(ctx, bson.M{"vessel_id": vesselID}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var events []domain.VesselEvent
	if err := cursor.All(ctx, &events); err != nil {
		return nil, err
	}
	return events, nil
}
