package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActivityPriority string

const (
	ActivityPriorityLow      ActivityPriority = "low"
	ActivityPriorityMedium   ActivityPriority = "medium"
	ActivityPriorityHigh     ActivityPriority = "high"
	ActivityPriorityCritical ActivityPriority = "critical"
)

type ActivityStatus string

const (
	ActivityStatusDraft       ActivityStatus = "draft"
	ActivityStatusSubmitted   ActivityStatus = "submitted"
	ActivityStatusApproved    ActivityStatus = "approved"
	ActivityStatusRejected    ActivityStatus = "rejected"
	ActivityStatusRescheduled ActivityStatus = "rescheduled"
	ActivityStatusSuspended   ActivityStatus = "suspended"
	ActivityStatusCompleted   ActivityStatus = "completed"
)

type Activity struct {
	ID               bson.ObjectID    `bson:"_id,omitempty" json:"id"`
	VesselID         bson.ObjectID    `bson:"vessel_id" json:"vessel_id"`
	Name             string           `bson:"name" json:"name"`
	Description      string           `bson:"description,omitempty" json:"description,omitempty"`
	StartDate        time.Time        `bson:"start_date" json:"start_date"`
	EndDate          time.Time        `bson:"end_date" json:"end_date"`
	DurationDays     int              `bson:"duration_days" json:"duration_days"`
	Priority         ActivityPriority `bson:"priority" json:"priority"`
	Status           ActivityStatus   `bson:"status" json:"status"`
	CreatedByUserID  bson.ObjectID    `bson:"created_by_user_id" json:"created_by_user_id"`
	ReviewedByUserID *bson.ObjectID   `bson:"reviewed_by_user_id,omitempty" json:"reviewed_by_user_id,omitempty"`
	ReviewNote       string           `bson:"review_note,omitempty" json:"review_note,omitempty"`
	CreatedAt        time.Time        `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time        `bson:"updated_at" json:"updated_at"`
}
