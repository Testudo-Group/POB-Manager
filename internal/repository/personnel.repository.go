package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrPersonnelNotFound = errors.New("personnel not found")

type PersonnelRepository struct {
	collection *mongo.Collection
}

func NewPersonnelRepository(db *mongo.Database) *PersonnelRepository {
	return &PersonnelRepository{
		collection: db.Collection("personnel"),
	}
}

func (r *PersonnelRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "employee_number", Value: 1},
			{Key: "company", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("personnel_empnum_company_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *PersonnelRepository) Create(ctx context.Context, p *domain.Personnel) error {
	_, err := r.collection.InsertOne(ctx, p)
	return err
}

func (r *PersonnelRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Personnel, error) {
	var p domain.Personnel
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&p)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrPersonnelNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PersonnelRepository) FindAll(ctx context.Context) ([]domain.Personnel, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"is_active": true})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.Personnel
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *PersonnelRepository) Update(ctx context.Context, p *domain.Personnel) error {
	filter := bson.M{"_id": p.ID}
	update := bson.M{"$set": p}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrPersonnelNotFound
	}
	
	return nil
}

func (r *PersonnelRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	// Soft delete
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_active": false}}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrPersonnelNotFound
	}
	
	return nil
}
