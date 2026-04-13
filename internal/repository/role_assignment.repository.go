package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrRoleAssignmentNotFound = errors.New("role assignment not found")

type RoleAssignmentRepository struct {
	collection *mongo.Collection
}

func NewRoleAssignmentRepository(db *mongo.Database) *RoleAssignmentRepository {
	return &RoleAssignmentRepository{
		collection: db.Collection("role_assignments"),
	}
}

func (r *RoleAssignmentRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "personnel_id", Value: 1},
			{Key: "offshore_role_id", Value: 1},
			{Key: "vessel_id", Value: 1},
		},
	})
	return err
}

func (r *RoleAssignmentRepository) Create(ctx context.Context, assignment *domain.RoleAssignment) error {
	_, err := r.collection.InsertOne(ctx, assignment)
	return err
}

func (r *RoleAssignmentRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.RoleAssignment, error) {
	var assignment domain.RoleAssignment
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&assignment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRoleAssignmentNotFound
		}
		return nil, err
	}
	return &assignment, nil
}

func (r *RoleAssignmentRepository) FindActiveByPersonnel(ctx context.Context, personnelID bson.ObjectID) ([]domain.RoleAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"personnel_id": personnelID,
		"status":       domain.RoleAssignmentStatusActive,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.RoleAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *RoleAssignmentRepository) FindActiveByVessel(ctx context.Context, vesselID bson.ObjectID) ([]domain.RoleAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"vessel_id": vesselID,
		"status":    domain.RoleAssignmentStatusActive,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.RoleAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *RoleAssignmentRepository) Update(ctx context.Context, assignment *domain.RoleAssignment) error {
	filter := bson.M{"_id": assignment.ID}
	update := bson.M{"$set": assignment}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrRoleAssignmentNotFound
	}
	return nil
}

func (r *RoleAssignmentRepository) CountActiveByRole(ctx context.Context, roleID, vesselID bson.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{
		"offshore_role_id": roleID,
		"vessel_id":        vesselID,
		"status":           domain.RoleAssignmentStatusActive,
	})
}