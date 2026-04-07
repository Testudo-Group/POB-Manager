package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type UserRole string

const (
	RoleActivityOwner UserRole = "activity_owner"
	RolePlanner       UserRole = "planner"
	RoleSafetyAdmin   UserRole = "safety_admin"
	RoleOIM           UserRole = "oim"
	RolePersonnel     UserRole = "personnel"
	RoleSystemAdmin   UserRole = "system_admin"
)

type User struct {
	ID                  bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	OrganizationID      bson.ObjectID  `bson:"organization_id" json:"organization_id"`
	FirstName           string         `bson:"first_name" json:"first_name"`
	LastName            string         `bson:"last_name" json:"last_name"`
	PhoneNumber         string         `bson:"phone_number" json:"phone_number"`
	Email               string         `bson:"email" json:"email"`
	PasswordHash        string         `bson:"password_hash" json:"-"`
	Role                UserRole       `bson:"role" json:"role"`
	VesselID            *bson.ObjectID `bson:"vessel_id,omitempty" json:"vessel_id,omitempty"`
	RefreshTokenHash    string         `bson:"refresh_token_hash,omitempty" json:"-"`
	RefreshTokenExpires *time.Time     `bson:"refresh_token_expires,omitempty" json:"-"`
	IsActive            bool           `bson:"is_active" json:"is_active"`
	CreatedAt           time.Time      `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time      `bson:"updated_at" json:"updated_at"`
}
