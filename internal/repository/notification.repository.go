package repository

import (
	"context"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type NotificationRepository struct {
	collection *mongo.Collection
}

func NewNotificationRepository(db *mongo.Database) *NotificationRepository {
	return &NotificationRepository{
		collection: db.Collection("notifications"),
	}
}

func (r *NotificationRepository) EnsureIndexes(ctx context.Context) error {
	index := mongo.IndexModel{
		Keys: bson.D{
			{Key: "user_id", Value: 1},
			{Key: "status", Value: 1},
		},
		Options: options.Index().SetName("notifications_user_id_status"),
	}

	_, err := r.collection.Indexes().CreateOne(ctx, index)
	return err
}

func (r *NotificationRepository) Create(ctx context.Context, n *domain.Notification) error {
	_, err := r.collection.InsertOne(ctx, n)
	return err
}

func (r *NotificationRepository) FindByUserID(ctx context.Context, userID bson.ObjectID) ([]domain.Notification, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var list []domain.Notification
	if err := cursor.All(ctx, &list); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *NotificationRepository) UpdateStatus(ctx context.Context, id bson.ObjectID, status domain.NotificationStatus) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$set": bson.M{"status": status}})
	return err
}
