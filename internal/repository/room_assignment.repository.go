package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrRoomAssignmentNotFound = errors.New("room assignment not found")

type RoomAssignmentRepository struct {
	collection *mongo.Collection
}

func NewRoomAssignmentRepository(db *mongo.Database) *RoomAssignmentRepository {
	return &RoomAssignmentRepository{
		collection: db.Collection("room_assignments"),
	}
}

func (r *RoomAssignmentRepository) EnsureIndexes(ctx context.Context) error {
	// Index on Vessel and Room for fast list queries
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "vessel_id", Value: 1},
			{Key: "room_id", Value: 1},
			{Key: "status", Value: 1},
		},
		Options: options.Index().SetName("room_assignments_lookup"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *RoomAssignmentRepository) Create(ctx context.Context, assignment *domain.RoomAssignment) error {
	_, err := r.collection.InsertOne(ctx, assignment)
	return err
}

func (r *RoomAssignmentRepository) FindActiveByRoom(ctx context.Context, roomID bson.ObjectID) ([]domain.RoomAssignment, error) {
	filter := bson.M{
		"room_id": roomID,
		"status":  domain.RoomAssignmentStatusActive,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.RoomAssignment
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *RoomAssignmentRepository) FindActiveByVessel(ctx context.Context, vesselID bson.ObjectID) ([]domain.RoomAssignment, error) {
	filter := bson.M{
		"vessel_id": vesselID,
		"status":    domain.RoomAssignmentStatusActive,
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.RoomAssignment
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}
