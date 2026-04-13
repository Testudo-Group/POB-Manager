package repository

import (
	"context"
	"errors"
	"time" // ADD THIS

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)
var ErrTravelAssignmentNotFound = errors.New("travel assignment not found")

type TravelAssignmentRepository struct {
	collection *mongo.Collection
}

func NewTravelAssignmentRepository(db *mongo.Database) *TravelAssignmentRepository {
	return &TravelAssignmentRepository{
		collection: db.Collection("travel_assignments"),
	}
}

func (r *TravelAssignmentRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{Keys: bson.D{{Key: "travel_schedule_id", Value: 1}}},
		{Keys: bson.D{{Key: "personnel_id", Value: 1}}},
	}
	for _, idx := range indexes {
		_, err := r.collection.Indexes().CreateOne(ctx, idx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *TravelAssignmentRepository) Create(ctx context.Context, assignment *domain.TravelAssignment) error {
	_, err := r.collection.InsertOne(ctx, assignment)
	return err
}

func (r *TravelAssignmentRepository) CreateMany(ctx context.Context, assignments []domain.TravelAssignment) error {
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

func (r *TravelAssignmentRepository) FindBySchedule(ctx context.Context, scheduleID bson.ObjectID) ([]domain.TravelAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"travel_schedule_id": scheduleID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.TravelAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *TravelAssignmentRepository) FindByPersonnel(ctx context.Context, personnelID bson.ObjectID) ([]domain.TravelAssignment, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"personnel_id": personnelID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var assignments []domain.TravelAssignment
	if err := cursor.All(ctx, &assignments); err != nil {
		return nil, err
	}
	return assignments, nil
}

func (r *TravelAssignmentRepository) CountBySchedule(ctx context.Context, scheduleID bson.ObjectID) (int64, error) {
	return r.collection.CountDocuments(ctx, bson.M{"travel_schedule_id": scheduleID})
}

func (r *TravelAssignmentRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.TravelAssignmentStatus) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"status": status, "updated_at": time.Now()}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrTravelAssignmentNotFound
	}
	return nil
}

func (r *TravelAssignmentRepository) DeleteBySchedule(ctx context.Context, scheduleID bson.ObjectID) error {
	_, err := r.collection.DeleteMany(ctx, bson.M{"travel_schedule_id": scheduleID})
	return err
}