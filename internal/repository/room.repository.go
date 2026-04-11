package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrRoomNotFound = errors.New("room not found")

type RoomRepository struct {
	collection *mongo.Collection
}

func NewRoomRepository(db *mongo.Database) *RoomRepository {
	return &RoomRepository{
		collection: db.Collection("rooms"),
	}
}

func (r *RoomRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "vessel_id", Value: 1},
			{Key: "code", Value: 1},
		},
		Options: options.Index().
			SetUnique(true).
			SetName("room_vessel_code_unique"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *RoomRepository) Create(ctx context.Context, room *domain.Room) error {
	_, err := r.collection.InsertOne(ctx, room)
	return err
}

func (r *RoomRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Room, error) {
	var room domain.Room
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&room)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRoomNotFound
		}
		return nil, err
	}

	return &room, nil
}

func (r *RoomRepository) FindByVessel(ctx context.Context, vesselID bson.ObjectID) ([]domain.Room, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"vessel_id": vesselID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.Room
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *RoomRepository) Update(ctx context.Context, room *domain.Room) error {
	filter := bson.M{"_id": room.ID}
	update := bson.M{"$set": room}
	
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	
	if result.MatchedCount == 0 {
		return ErrRoomNotFound
	}
	
	return nil
}

func (r *RoomRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	
	if result.DeletedCount == 0 {
		return ErrRoomNotFound
	}
	
	return nil
}
