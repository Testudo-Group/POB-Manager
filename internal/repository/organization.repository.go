package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrOrganizationNotFound = errors.New("organization not found")

type OrganizationRepository struct {
	collection *mongo.Collection
}

func NewOrganizationRepository(db *mongo.Database) *OrganizationRepository {
	return &OrganizationRepository{
		collection: db.Collection("organizations"),
	}
}

func (r *OrganizationRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "name", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("organizations_name_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *OrganizationRepository) Create(ctx context.Context, organization *domain.Organization) error {
	_, err := r.collection.InsertOne(ctx, organization)
	return err
}

func (r *OrganizationRepository) FindByName(ctx context.Context, name string) (*domain.Organization, error) {
	var organization domain.Organization
	err := r.collection.FindOne(ctx, bson.M{"name": name}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	return &organization, nil
}

func (r *OrganizationRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Organization, error) {
	var organization domain.Organization
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&organization)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOrganizationNotFound
		}
		return nil, err
	}

	return &organization, nil
}
