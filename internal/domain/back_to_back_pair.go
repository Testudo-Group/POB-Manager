package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type BackToBackPairStatus string

const (
	BackToBackPairStatusActive   BackToBackPairStatus = "active"
	BackToBackPairStatusInactive BackToBackPairStatus = "inactive"
)

type BackToBackPair struct {
	ID                 bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	OffshoreRoleID     bson.ObjectID        `bson:"offshore_role_id" json:"offshore_role_id"`
	VesselID           bson.ObjectID        `bson:"vessel_id" json:"vessel_id"`
	PrimaryPersonnelID bson.ObjectID        `bson:"primary_personnel_id" json:"primary_personnel_id"`
	ReliefPersonnelID  bson.ObjectID        `bson:"relief_personnel_id" json:"relief_personnel_id"`
	RoomID             *bson.ObjectID       `bson:"room_id,omitempty" json:"room_id,omitempty"`
	Notes              string               `bson:"notes,omitempty" json:"notes,omitempty"`
	Status             BackToBackPairStatus `bson:"status" json:"status"`
	EffectiveFrom      time.Time            `bson:"effective_from" json:"effective_from"`
	EffectiveTo        *time.Time           `bson:"effective_to,omitempty" json:"effective_to,omitempty"`
	CreatedAt          time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time            `bson:"updated_at" json:"updated_at"`
}
