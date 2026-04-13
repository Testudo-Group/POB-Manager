package domain

import (
	"time"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TravelAssignmentStatus string

const (
	TravelAssignmentStatusReserved  TravelAssignmentStatus = "reserved"
	TravelAssignmentStatusConfirmed TravelAssignmentStatus = "confirmed"
	TravelAssignmentStatusBoarded   TravelAssignmentStatus = "boarded"
	TravelAssignmentStatusCancelled TravelAssignmentStatus = "cancelled"
)

type TravelAssignment struct {
	ID               bson.ObjectID          `bson:"_id,omitempty" json:"id"`
	TravelScheduleID bson.ObjectID          `bson:"travel_schedule_id" json:"travel_schedule_id"`
	PersonnelID      bson.ObjectID          `bson:"personnel_id" json:"personnel_id"`
	Status           TravelAssignmentStatus `bson:"status" json:"status"`
	AssignedAt       time.Time              `bson:"assigned_at" json:"assigned_at"`
	CreatedAt        time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time              `bson:"updated_at" json:"updated_at"`
}
