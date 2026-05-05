package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type VesselType string

const (
	VesselTypePrimary      VesselType = "primary"
	VesselTypeSecondary    VesselType = "secondary"
	VesselTypeInstallation VesselType = "installation"
)

type VesselStatus string

const (
	VesselStatusActive   VesselStatus = "active"
	VesselStatusInactive VesselStatus = "inactive"
)

type Vessel struct {
	ID                     bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                   string        `bson:"name" json:"name"`
	Code                   string        `bson:"code" json:"code"`
	Type                   VesselType    `bson:"type" json:"type"`
	Location               string        `bson:"location" json:"location"`
	POBCapacity            int           `bson:"pob_capacity" json:"pob_capacity"`
	MinimumSafePOBCapacity int           `bson:"minimum_safe_pob_capacity" json:"minimum_safe_pob_capacity"`
	IsMinimumManningActive bool          `bson:"is_minimum_manning_active" json:"is_minimum_manning_active"`
	Decks                  []string      `bson:"decks,omitempty" json:"decks,omitempty"`
	IsDefault              bool          `bson:"is_default" json:"is_default"`
	Status                 VesselStatus  `bson:"status" json:"status"`
	CreatedAt              time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt              time.Time     `bson:"updated_at" json:"updated_at"`
}
