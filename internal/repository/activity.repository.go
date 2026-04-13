package repository

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var ErrActivityNotFound = errors.New("activity not found")

type ActivityRepository struct {
	collection *mongo.Collection
}

func NewActivityRepository(db *mongo.Database) *ActivityRepository {
	return &ActivityRepository{
		collection: db.Collection("activities"),
	}
}

func (r *ActivityRepository) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "vessel_id", Value: 1}, {Key: "status", Value: 1}},
			Options: options.Index().SetName("activities_vessel_status"),
		},
		{
			Keys:    bson.D{{Key: "start_date", Value: 1}, {Key: "end_date", Value: 1}},
			Options: options.Index().SetName("activities_dates"),
		},
		{
			Keys:    bson.D{{Key: "created_by_user_id", Value: 1}},
			Options: options.Index().SetName("activities_created_by"),
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

func (r *ActivityRepository) Create(ctx context.Context, activity *domain.Activity) error {
	_, err := r.collection.InsertOne(ctx, activity)
	return err
}

func (r *ActivityRepository) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Activity, error) {
	var activity domain.Activity
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&activity)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrActivityNotFound
		}
		return nil, err
	}
	return &activity, nil
}

func (r *ActivityRepository) FindAll(ctx context.Context, filter bson.M) ([]domain.Activity, error) {
	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var activities []domain.Activity
	if err := cursor.All(ctx, &activities); err != nil {
		return nil, err
	}
	return activities, nil
}

func (r *ActivityRepository) FindByVessel(ctx context.Context, vesselID bson.ObjectID, statuses ...domain.ActivityStatus) ([]domain.Activity, error) {
	filter := bson.M{"vessel_id": vesselID}
	if len(statuses) > 0 {
		filter["status"] = bson.M{"$in": statuses}
	}
	return r.FindAll(ctx, filter)
}

func (r *ActivityRepository) FindByDateRange(ctx context.Context, vesselID bson.ObjectID, startDate, endDate time.Time) ([]domain.Activity, error) {
	filter := bson.M{
		"vessel_id": vesselID,
		"$or": []bson.M{
			{"start_date": bson.M{"$gte": startDate, "$lte": endDate}},
			{"end_date": bson.M{"$gte": startDate, "$lte": endDate}},
			{"start_date": bson.M{"$lte": startDate}, "end_date": bson.M{"$gte": endDate}},
		},
	}
	return r.FindAll(ctx, filter)
}

func (r *ActivityRepository) FindPendingApproval(ctx context.Context) ([]domain.Activity, error) {
	filter := bson.M{"status": domain.ActivityStatusSubmitted}
	return r.FindAll(ctx, filter)
}

func (r *ActivityRepository) Update(ctx context.Context, activity *domain.Activity) error {
	filter := bson.M{"_id": activity.ID}
	update := bson.M{"$set": activity}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *ActivityRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.ActivityStatus, reviewedBy *bson.ObjectID, note string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":             status,
			"reviewed_by_user_id": reviewedBy,
			"review_note":        note,
			"updated_at":         time.Now(),
		},
	}
	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return ErrActivityNotFound
	}
	return nil
}

func (r *ActivityRepository) Delete(ctx context.Context, id bson.ObjectID) error {
	result, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if result.DeletedCount == 0 {
		return ErrActivityNotFound
	}
	return nil
}