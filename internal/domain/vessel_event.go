package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type VesselEventType string

const (
	VesselEventTypeLocationChange VesselEventType = "location_change"
	VesselEventTypeStatusChange   VesselEventType = "status_change"
	VesselEventTypeMaintenance    VesselEventType = "maintenance"
	VesselEventTypeDeployment     VesselEventType = "deployment"
	VesselEventTypeReturn         VesselEventType = "return"
	VesselEventTypeInspection     VesselEventType = "inspection"
	VesselEventTypeNote           VesselEventType = "note"
)

type VesselEvent struct {
	ID               bson.ObjectID   `bson:"_id,omitempty" json:"id"`
	VesselID         bson.ObjectID   `bson:"vessel_id" json:"vessel_id"`
	EventType        VesselEventType `bson:"event_type" json:"event_type"`
	Description      string          `bson:"description" json:"description"`
	Location         string          `bson:"location,omitempty" json:"location,omitempty"`
	RecordedByUserID *bson.ObjectID  `bson:"recorded_by_user_id,omitempty" json:"recorded_by_user_id,omitempty"`
	OccurredAt       time.Time       `bson:"occurred_at" json:"occurred_at"`
	CreatedAt        time.Time       `bson:"created_at" json:"created_at"`
}
