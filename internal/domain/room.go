package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomStatus string

const (
	RoomStatusAvailable   RoomStatus = "available"
	RoomStatusMaintenance RoomStatus = "maintenance"
	RoomStatusInactive    RoomStatus = "inactive"
)

type Room struct {
	ID               bson.ObjectID `bson:"_id,omitempty" json:"id"`
	VesselID         bson.ObjectID `bson:"vessel_id" json:"vessel_id"`
	Name             string        `bson:"name" json:"room_name"`
	Code             string        `bson:"code" json:"room_code"`
	Deck             string        `bson:"deck" json:"location_deck"`
	Category         string        `bson:"category" json:"room_category"`
	Capacity         int           `bson:"capacity" json:"capacity"`
	Description      string        `bson:"description,omitempty" json:"description,omitempty"`
	Status           RoomStatus    `bson:"status" json:"status"`
	CurrentOccupancy int           `bson:"-" json:"current_occupancy"`
	CreatedAt        time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time     `bson:"updated_at" json:"updated_at"`
}
