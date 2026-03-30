package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationType string

const (
	NotificationTypeComplianceExpiry NotificationType = "compliance_expiry"
	NotificationTypeTravelBlocked    NotificationType = "travel_blocked"
	NotificationTypeMinimumManning   NotificationType = "minimum_manning"
	NotificationTypeActivityReview   NotificationType = "activity_review"
	NotificationTypeTravelAssignment NotificationType = "travel_assignment"
)

type NotificationChannel string

const (
	NotificationChannelInApp NotificationChannel = "in_app"
	NotificationChannelEmail NotificationChannel = "email"
)

type NotificationStatus string

const (
	NotificationStatusPending NotificationStatus = "pending"
	NotificationStatusSent    NotificationStatus = "sent"
	NotificationStatusRead    NotificationStatus = "read"
	NotificationStatusFailed  NotificationStatus = "failed"
)

type Notification struct {
	ID                bson.ObjectID       `bson:"_id,omitempty" json:"id"`
	UserID            *bson.ObjectID      `bson:"user_id,omitempty" json:"user_id,omitempty"`
	PersonnelID       *bson.ObjectID      `bson:"personnel_id,omitempty" json:"personnel_id,omitempty"`
	Type              NotificationType    `bson:"type" json:"type"`
	Channel           NotificationChannel `bson:"channel" json:"channel"`
	Title             string              `bson:"title" json:"title"`
	Message           string              `bson:"message" json:"message"`
	RelatedEntityType string              `bson:"related_entity_type,omitempty" json:"related_entity_type,omitempty"`
	RelatedEntityID   *bson.ObjectID      `bson:"related_entity_id,omitempty" json:"related_entity_id,omitempty"`
	Status            NotificationStatus  `bson:"status" json:"status"`
	SentAt            *time.Time          `bson:"sent_at,omitempty" json:"sent_at,omitempty"`
	ReadAt            *time.Time          `bson:"read_at,omitempty" json:"read_at,omitempty"`
	CreatedAt         time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time           `bson:"updated_at" json:"updated_at"`
}
