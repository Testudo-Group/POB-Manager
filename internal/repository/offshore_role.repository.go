package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrOffshoreRoleNotFound = errors.New("offshore role not found")

type OffshoreRoleRepository struct {
	collection *mongo.Collection
}

func NewOffshoreRoleRepository(db *mongo.Database) *OffshoreRoleRepository {
	return &OffshoreRoleRepository{
		collection: db.Collection("offshore_roles"),
	}
}

func (r *OffshoreRoleRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{{Key: "code", Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("offshore_roles_code_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *OffshoreRoleRepository) Create(ctx context.Context, role *domain.OffshoreRole) error {
	_, err := r.collection.InsertOne(ctx, role)
	return err
}

func (r *OffshoreRoleRepository) Update(ctx context.Context, role *domain.OffshoreRole) error {
	filter := bson.M{"_id": role.ID}
	update := bson.M{"$set": role}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrOffshoreRoleNotFound
	}
	
	return nil
}

func (r *OffshoreRoleRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.OffshoreRole, error) {
	var role domain.OffshoreRole
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&role)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrOffshoreRoleNotFound
		}
		return nil, err
	}

	return &role, nil
}

func (r *OffshoreRoleRepository) FindAll(ctx context.Context) ([]domain.OffshoreRole, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var roles []domain.OffshoreRole
	if err := cursor.All(ctx, &roles); err != nil {
		return nil, err
	}

	return roles, nil
}
