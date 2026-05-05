package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OffshoreRoleType string

const (
	OffshoreRoleTypeCore     OffshoreRoleType = "core"
	OffshoreRoleTypeFlexible OffshoreRoleType = "flexible"
)

type OffshoreRoleStatus string

const (
	OffshoreRoleStatusActive   OffshoreRoleStatus = "active"
	OffshoreRoleStatusInactive OffshoreRoleStatus = "inactive"
)

type OffshoreRole struct {
	ID                       bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	Name                     string             `bson:"name" json:"name"`
	Code                     string             `bson:"code" json:"code"`
	Description              string             `bson:"description,omitempty" json:"description,omitempty"`
	Type                     OffshoreRoleType   `bson:"type" json:"type"`
	VesselID                 *bson.ObjectID     `bson:"vessel_id,omitempty" json:"vessel_id,omitempty"`
	RequiresRoom             bool               `bson:"requires_room" json:"requires_room"`
	MinimumRequiredCount     int                `bson:"minimum_required_count" json:"minimum_required_count"`
	Status                   OffshoreRoleStatus `bson:"status" json:"status"`
	RequiredCertificateTypes []string           `bson:"required_certificate_types" json:"required_certificate_types"`
	CreatedAt                time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt                time.Time          `bson:"updated_at" json:"updated_at"`
}
