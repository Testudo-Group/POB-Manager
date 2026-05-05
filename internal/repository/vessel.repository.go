package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrVesselNotFound = errors.New("vessel not found")

type VesselRepository struct {
	collection *mongo.Collection
}

func NewVesselRepository(db *mongo.Database) *VesselRepository {
	return &VesselRepository{
		collection: db.Collection("vessels"),
	}
}

func (r *VesselRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "code", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("vessel_code_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *VesselRepository) Create(ctx context.Context, vessel *domain.Vessel) error {
	_, err := r.collection.InsertOne(ctx, vessel)
	return err
}

func (r *VesselRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Vessel, error) {
	var v domain.Vessel
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrVesselNotFound
		}
		return nil, err
	}

	return &v, nil
}

func (r *VesselRepository) FindAll(ctx context.Context) ([]domain.Vessel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"status": domain.VesselStatusActive})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.Vessel
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *VesselRepository) Update(ctx context.Context, v *domain.Vessel) error {
	filter := bson.M{"_id": v.ID}
	update := bson.M{"$set": v}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrVesselNotFound
	}
	
	return nil
}

func (r *VesselRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": domain.VesselStatusInactive}}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return ErrVesselNotFound
	}

	return nil
}

func (r *VesselRepository) FindDefault(ctx context.Context) (*domain.Vessel, error) {
	var v domain.Vessel
	err := r.collection.FindOne(ctx, bson.M{
		"is_default": true,
		"status":     domain.VesselStatusActive,
	}).Decode(&v)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Fall back to first active vessel
			cursor, findErr := r.collection.Find(ctx, bson.M{"status": domain.VesselStatusActive},
				options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}}).SetLimit(1))
			if findErr != nil {
				return nil, findErr
			}
			defer cursor.Close(ctx)
			if cursor.Next(ctx) {
				if decodeErr := cursor.Decode(&v); decodeErr != nil {
					return nil, decodeErr
				}
				return &v, nil
			}
			return nil, ErrVesselNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *VesselRepository) SetDefault(ctx context.Context, id bson.ObjectID) error {
	// Clear existing default
	_, err := r.collection.UpdateMany(ctx,
		bson.M{"is_default": true},
		bson.M{"$set": bson.M{"is_default": false}},
	)
	if err != nil {
		return err
	}

	result, err := r.collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"is_default": true}},
	)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrVesselNotFound
	}
	return nil
}
