package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TravelAssignmentStatus string

const (
	TravelAssignmentStatusProposed  TravelAssignmentStatus = "proposed"
	TravelAssignmentStatusConfirmed TravelAssignmentStatus = "confirmed"
	TravelAssignmentStatusBoarded   TravelAssignmentStatus = "boarded"
	TravelAssignmentStatusCancelled TravelAssignmentStatus = "cancelled"
	TravelAssignmentStatusBlocked   TravelAssignmentStatus = "blocked"
)

type TravelAssignment struct {
	ID                  bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	TravelScheduleID    bson.ObjectID          `bson:"travel_schedule_id" json:"travel_schedule_id"`
	PersonnelID         bson.ObjectID          `bson:"personnel_id" json:"personnel_id"`
	ActivityID          *bson.ObjectID         `bson:"activity_id,omitempty" json:"activity_id,omitempty"`
	ApprovedByUserID    *bson.ObjectID         `bson:"approved_by_user_id,omitempty" json:"approved_by_user_id,omitempty"`
	ComplianceCheckedAt *time.Time             `bson:"compliance_checked_at,omitempty" json:"compliance_checked_at,omitempty"`
	BlockReason         string                 `bson:"block_reason,omitempty" json:"block_reason,omitempty"`
	Status              TravelAssignmentStatus `bson:"status" json:"status"`
	CreatedAt           time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `bson:"updated_at" json:"updated_at"`
}
