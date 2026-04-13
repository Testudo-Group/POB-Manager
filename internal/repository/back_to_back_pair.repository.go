package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrBackToBackPairNotFound = errors.New("back-to-back pair not found")

type BackToBackPairRepository struct {
	collection *mongo.Collection
}

func NewBackToBackPairRepository(db *mongo.Database) *BackToBackPairRepository {
	return &BackToBackPairRepository{
		collection: db.Collection("back_to_back_pairs"),
	}
}

func (r *BackToBackPairRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "offshore_role_id", Value: 1},
			{Key: "vessel_id", Value: 1},
			{Key: "status", Value: 1},
		},
	})
	return err
}

func (r *BackToBackPairRepository) Create(ctx context.Context, pair *domain.BackToBackPair) error {
	_, err := r.collection.InsertOne(ctx, pair)
	return err
}

func (r *BackToBackPairRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.BackToBackPair, error) {
	var pair domain.BackToBackPair
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&pair)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrBackToBackPairNotFound
		}
		return nil, err
	}
	return &pair, nil
}

func (r *BackToBackPairRepository) FindActiveByRole(ctx context.Context, roleID, vesselID bson.ObjectID) ([]domain.BackToBackPair, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"offshore_role_id": roleID,
		"vessel_id":        vesselID,
		"status":           domain.BackToBackPairStatusActive,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pairs []domain.BackToBackPair
	if err := cursor.All(ctx, &pairs); err != nil {
		return nil, err
	}
	return pairs, nil
}

func (r *BackToBackPairRepository) FindByPersonnel(ctx context.Context, personnelID bson.ObjectID) ([]domain.BackToBackPair, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"$or": []bson.M{
			{"primary_personnel_id": personnelID},
			{"relief_personnel_id": personnelID},
		},
		"status": domain.BackToBackPairStatusActive,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var pairs []domain.BackToBackPair
	if err := cursor.All(ctx, &pairs); err != nil {
		return nil, err
	}
	return pairs, nil
}

func (r *BackToBackPairRepository) Update(ctx context.Context, pair *domain.BackToBackPair) error {
	filter := bson.M{"_id": pair.ID}
	update := bson.M{"$set": pair}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBackToBackPairNotFound
	}
	return nil
}

func (r *BackToBackPairRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": domain.BackToBackPairStatusInactive}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrBackToBackPairNotFound
	}
	return nil
}