package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TransportType string

const (
	TransportTypeHelicopter TransportType = "helicopter"
	TransportTypeBoat       TransportType = "boat"
	TransportTypePickup     TransportType = "pickup"
	TransportTypeHiace      TransportType = "hiace"
)

type TransportCostModel string

const (
	TransportCostModelPerTrip TransportCostModel = "per_trip"
	TransportCostModelPerSeat TransportCostModel = "per_seat"
)

type Transport struct {
	ID                   bson.ObjectID      `bson:"_id,omitempty" json:"id"`
	Name                 string             `bson:"name" json:"name"`
	Type                 TransportType      `bson:"type" json:"type"`
	Capacity             int                `bson:"capacity" json:"capacity"`
	CostModel            TransportCostModel `bson:"cost_model" json:"cost_model"`
	CostAmount           float64            `bson:"cost_amount" json:"cost_amount"`
	DepartureDays        []string           `bson:"departure_days" json:"departure_days"`
	MobilizationLocation string             `bson:"mobilization_location" json:"mobilization_location"`
	IsActive             bool               `bson:"is_active" json:"is_active"`
	CreatedAt            time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time          `bson:"updated_at" json:"updated_at"`
}
