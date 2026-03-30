package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoleAssignmentStatus string

const (
	RoleAssignmentStatusPlanned   RoleAssignmentStatus = "planned"
	RoleAssignmentStatusActive    RoleAssignmentStatus = "active"
	RoleAssignmentStatusRelief    RoleAssignmentStatus = "relief"
	RoleAssignmentStatusCompleted RoleAssignmentStatus = "completed"
	RoleAssignmentStatusCancelled RoleAssignmentStatus = "cancelled"
)

type RoleAssignment struct {
	ID                 bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	OffshoreRoleID     bson.ObjectID        `bson:"offshore_role_id" json:"offshore_role_id"`
	PersonnelID        bson.ObjectID        `bson:"personnel_id" json:"personnel_id"`
	VesselID           bson.ObjectID        `bson:"vessel_id" json:"vessel_id"`
	RotationScheduleID *bson.ObjectID       `bson:"rotation_schedule_id,omitempty" json:"rotation_schedule_id,omitempty"`
	RoomID             *bson.ObjectID       `bson:"room_id,omitempty" json:"room_id,omitempty"`
	AssignedByUserID   *bson.ObjectID       `bson:"assigned_by_user_id,omitempty" json:"assigned_by_user_id,omitempty"`
	EffectiveFrom      time.Time            `bson:"effective_from" json:"effective_from"`
	EffectiveTo        *time.Time           `bson:"effective_to,omitempty" json:"effective_to,omitempty"`
	Status             RoleAssignmentStatus `bson:"status" json:"status"`
	CreatedAt          time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time            `bson:"updated_at" json:"updated_at"`
}
