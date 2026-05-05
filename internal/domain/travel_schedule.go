package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TravelDirection string

const (
	TravelDirectionInbound  TravelDirection = "inbound"
	TravelDirectionOutbound TravelDirection = "outbound"
)

type TravelScheduleStatus string

const (
	TravelScheduleStatusPlanned   TravelScheduleStatus = "planned"
	TravelScheduleStatusConfirmed TravelScheduleStatus = "confirmed"
	TravelScheduleStatusDeparted  TravelScheduleStatus = "departed"
	TravelScheduleStatusCompleted TravelScheduleStatus = "completed"
	TravelScheduleStatusCancelled TravelScheduleStatus = "cancelled"
)

type TravelSchedule struct {
	ID                  bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	TransportID         bson.ObjectID        `bson:"transport_id" json:"transport_id"`
	VesselID            *bson.ObjectID       `bson:"vessel_id,omitempty" json:"vessel_id,omitempty"`
	OriginVesselID      *bson.ObjectID       `bson:"origin_vessel_id,omitempty" json:"origin_vessel_id,omitempty"`
	DestinationVesselID *bson.ObjectID       `bson:"destination_vessel_id,omitempty" json:"destination_vessel_id,omitempty"`
	ActivityID          *bson.ObjectID       `bson:"activity_id,omitempty" json:"activity_id,omitempty"`
	Direction           TravelDirection      `bson:"direction" json:"direction"`
	DepartureAt         time.Time            `bson:"departure_at" json:"departure_at"`
	ArrivalAt           *time.Time           `bson:"arrival_at,omitempty" json:"arrival_at,omitempty"`
	SeatCapacity        int                  `bson:"seat_capacity" json:"seat_capacity"`
	ReservedSeats       int                  `bson:"reserved_seats" json:"reserved_seats"`
	Status              TravelScheduleStatus `bson:"status" json:"status"`
	CreatedAt           time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt           time.Time            `bson:"updated_at" json:"updated_at"`
}
