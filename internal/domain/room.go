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
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	VesselID  bson.ObjectID `bson:"vessel_id" json:"vessel_id"`
	Name      string        `bson:"name" json:"name"`
	Code      string        `bson:"code" json:"code"`
	Deck      string        `bson:"deck" json:"deck"`
	Category  string        `bson:"category" json:"category"`
	Capacity  int           `bson:"capacity" json:"capacity"`
	Status    RoomStatus    `bson:"status" json:"status"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}
