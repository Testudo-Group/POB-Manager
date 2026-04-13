package repository

import (
	"context"
	"errors"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ErrRotationScheduleNotFound = errors.New("rotation schedule not found")

type RotationScheduleRepository struct {
	collection *mongo.Collection
}

func NewRotationScheduleRepository(db *mongo.Database) *RotationScheduleRepository {
	return &RotationScheduleRepository{
		collection: db.Collection("rotation_schedules"),
	}
}

func (r *RotationScheduleRepository) EnsureIndexes(ctx context.Context) error {
	_, err := r.collection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "offshore_role_id", Value: 1}, {Key: "vessel_id", Value: 1}},
	})
	return err
}

func (r *RotationScheduleRepository) Create(ctx context.Context, schedule *domain.RotationSchedule) error {
	_, err := r.collection.InsertOne(ctx, schedule)
	return err
}

func (r *RotationScheduleRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.RotationSchedule, error) {
	var schedule domain.RotationSchedule
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&schedule)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrRotationScheduleNotFound
		}
		return nil, err
	}
	return &schedule, nil
}

func (r *RotationScheduleRepository) FindByRoleAndVessel(ctx context.Context, roleID, vesselID bson.ObjectID) ([]domain.RotationSchedule, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"offshore_role_id": roleID,
		"vessel_id":        vesselID,
		"is_active":        true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var schedules []domain.RotationSchedule
	if err := cursor.All(ctx, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func (r *RotationScheduleRepository) Update(ctx context.Context, schedule *domain.RotationSchedule) error {
	filter := bson.M{"_id": schedule.ID}
	update := bson.M{"$set": schedule}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrRotationScheduleNotFound
	}
	return nil
}

func (r *RotationScheduleRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	filter := bson.M{"_id": id}
	update := bson.M{"$set": bson.M{"is_active": false}}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrRotationScheduleNotFound
	}
	return nil
}