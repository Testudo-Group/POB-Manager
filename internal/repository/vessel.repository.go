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
	cursor, err := r.collection.Find(ctx, bson.M{})
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
