package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PersonnelStatus string

const (
	PersonnelStatusAvailable         PersonnelStatus = "available"
	PersonnelStatusOnboard           PersonnelStatus = "onboard"
	PersonnelStatusTravelling        PersonnelStatus = "travelling"
	PersonnelStatusOffboard          PersonnelStatus = "offboard"
	PersonnelStatusComplianceBlocked PersonnelStatus = "compliance_blocked"
	PersonnelStatusInactive          PersonnelStatus = "inactive"
)

type Personnel struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	UserID            *bson.ObjectID    `bson:"user_id,omitempty" json:"user_id,omitempty"`
	EmployeeNumber    string            `bson:"employee_number" json:"employee_number"`
	FirstName         string            `bson:"first_name" json:"first_name"`
	LastName          string            `bson:"last_name" json:"last_name"`
	Email             string            `bson:"email" json:"email"`
	PhoneNumber       string            `bson:"phone_number" json:"phone_number"`
	Nationality       string            `bson:"nationality" json:"nationality"`
	Company           string            `bson:"company" json:"company"`
	PrimaryDiscipline string            `bson:"primary_discipline" json:"primary_discipline"`
	OffshoreRoleIDs   []bson.ObjectID   `bson:"offshore_role_ids" json:"offshore_role_ids"`
	CurrentStatus     PersonnelStatus   `bson:"current_status" json:"current_status"`
	CurrentVesselID   *bson.ObjectID    `bson:"current_vessel_id,omitempty" json:"current_vessel_id,omitempty"`
	IsActive          bool              `bson:"is_active" json:"is_active"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
}
