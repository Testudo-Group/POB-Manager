package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomAssignmentStatus string

const (
	RoomAssignmentStatusReserved RoomAssignmentStatus = "reserved"
	RoomAssignmentStatusActive   RoomAssignmentStatus = "active"
	RoomAssignmentStatusReleased RoomAssignmentStatus = "released"
)

type RoomAssignment struct {
	ID               bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	VesselID         bson.ObjectID        `bson:"vessel_id" json:"vessel_id"`
	RoomID           bson.ObjectID        `bson:"room_id" json:"room_id"`
	PersonnelID      bson.ObjectID        `bson:"personnel_id" json:"personnel_id"`
	OffshoreRoleID   *bson.ObjectID       `bson:"offshore_role_id,omitempty" json:"offshore_role_id,omitempty"`
	RoleAssignmentID *bson.ObjectID       `bson:"role_assignment_id,omitempty" json:"role_assignment_id,omitempty"`
	ActivityID       *bson.ObjectID       `bson:"activity_id,omitempty" json:"activity_id,omitempty"`
	StartsAt         time.Time            `bson:"starts_at" json:"starts_at"`
	EndsAt           *time.Time           `bson:"ends_at,omitempty" json:"ends_at,omitempty"`
	Status           RoomAssignmentStatus `bson:"status" json:"status"`
	CreatedAt        time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time            `bson:"updated_at" json:"updated_at"`
}
