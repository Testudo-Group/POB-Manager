package repository

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrActivityAssignmentNotFound = errors.New("activity assignment not found")

type ActivityAssignmentRepository struct {
	collection *mongo.Collection
}

func NewActivityAssignmentRepository(db *mongo.Database) *ActivityAssignmentRepository {
	return &ActivityAssignmentRepository{
		collection: db.Collection("activity_assignments"),
	}
}

func (r *ActivityAssignmentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "activity_id", Value: 1}, {Key: "personnel_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "personnel_id", Value: 1}, {Key: "status", Value: 1}},
		},
	}

	for _, idx := range indexes {
		_, err := r.collection.Indexes().CreateOne(ctx, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ActivityAssignmentRepository) Create(ctx context.Context, assignment *domain.ActivityAssignment) error {
	_, err := r.collection.InsertOne(ctx, assignment)
	return err
}

func (r *ActivityAssignmentRepository) CreateMany(ctx context.Context, assignments []domain.ActivityAssignment) error {
	if len(assignments) == 0 {
		return nil
	}
	docs := make([]interface{}, len(assignments))
	for i, a := range assignments {
		docs[i] = a
	}
	_, err := r.collection.InsertMany(ctx, docs)
	return err
}

func (r *ActivityAssignmentRepository) FindByActivity(ctx context.Context, activityID bson.ObjectID) ([]domain.ActivityAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"activity_id": activityID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.ActivityAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *ActivityAssignmentRepository) FindByPersonnel(ctx context.Context, personnelID bson.ObjectID) ([]domain.ActivityAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"personnel_id": personnelID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.ActivityAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *ActivityAssignmentRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.ActivityAssignmentStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrActivityAssignmentNotFound
	}
	return nil
}

func (r *ActivityAssignmentRepository) DeleteByActivity(ctx context.Context, activityID bson.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"activity_id": activityID})
	return err
}