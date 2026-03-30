package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AssignmentComplianceStatus string

const (
	AssignmentComplianceStatusPending AssignmentComplianceStatus = "pending"
	AssignmentComplianceStatusValid   AssignmentComplianceStatus = "valid"
	AssignmentComplianceStatusBlocked AssignmentComplianceStatus = "blocked"
)

type ActivityAssignmentStatus string

const (
	ActivityAssignmentStatusPlanned   ActivityAssignmentStatus = "planned"
	ActivityAssignmentStatusConfirmed ActivityAssignmentStatus = "confirmed"
	ActivityAssignmentStatusCancelled ActivityAssignmentStatus = "cancelled"
	ActivityAssignmentStatusCompleted ActivityAssignmentStatus = "completed"
)

type ActivityAssignment struct {
	ID               bson.ObjectID              `bson:"_id,omitempty" json:"id"`
	ActivityID       bson.ObjectID              `bson:"activity_id" json:"activity_id"`
	PersonnelID      bson.ObjectID              `bson:"personnel_id" json:"personnel_id"`
	OffshoreRoleID   bson.ObjectID              `bson:"offshore_role_id" json:"offshore_role_id"`
	RoleAssignmentID *bson.ObjectID             `bson:"role_assignment_id,omitempty" json:"role_assignment_id,omitempty"`
	TravelScheduleID *bson.ObjectID             `bson:"travel_schedule_id,omitempty" json:"travel_schedule_id,omitempty"`
	RoomAssignmentID *bson.ObjectID             `bson:"room_assignment_id,omitempty" json:"room_assignment_id,omitempty"`
	ComplianceStatus AssignmentComplianceStatus `bson:"compliance_status" json:"compliance_status"`
	Status           ActivityAssignmentStatus   `bson:"status" json:"status"`
	AssignedByUserID *bson.ObjectID             `bson:"assigned_by_user_id,omitempty" json:"assigned_by_user_id,omitempty"`
	CreatedAt        time.Time                  `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time                  `bson:"updated_at" json:"updated_at"`
}
