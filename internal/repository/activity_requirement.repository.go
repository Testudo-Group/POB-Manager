package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrActivityRequirementNotFound = errors.New("activity requirement not found")

type ActivityRequirementRepository struct {
	collection *mongo.Collection
}

func NewActivityRequirementRepository(db *mongo.Database) *ActivityRequirementRepository {
	return &ActivityRequirementRepository{
		collection: db.Collection("activity_requirements"),
	}
}

func (r *ActivityRequirementRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "activity_id", Value: 1}},
	})
	return err
}

func (r *ActivityRequirementRepository) Create(ctx context.Context, req *domain.ActivityRequirement) error {
	_, err := r.collection.InsertOne(ctx, req)
	return err
}

func (r *ActivityRequirementRepository) CreateMany(ctx context.Context, reqs []domain.ActivityRequirement) error {
	if len(reqs) == 0 {
		return nil
	}
	docs := make([]interface{}, len(reqs))
	for i, req := range reqs {
		docs[i] = req
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *ActivityRequirementRepository) FindByActivity(ctx context.Context, activityID bson.ObjectID) ([]domain.ActivityRequirement, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"activity_id": activityID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var reqs []domain.ActivityRequirement
	if err := cursor.All(ctx, &reqs); err != nil {
		return nil, err
	}
	return reqs, nil
}

func (r *ActivityRequirementRepository) DeleteByActivity(ctx context.Context, activityID bson.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"activity_id": activityID})
	return err
}