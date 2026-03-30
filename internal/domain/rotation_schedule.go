package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type RotationSchedule struct {
	ID              bson.ObjectID `bson:"_id,omitempty" json:"id"`
	OffshoreRoleID  bson.ObjectID `bson:"offshore_role_id" json:"offshore_role_id"`
	VesselID        bson.ObjectID `bson:"vessel_id" json:"vessel_id"`
	Name            string        `bson:"name" json:"name"`
	DaysOn          int           `bson:"days_on" json:"days_on"`
	DaysOff         int           `bson:"days_off" json:"days_off"`
	CycleAnchorDate time.Time     `bson:"cycle_anchor_date" json:"cycle_anchor_date"`
	IsActive        bool          `bson:"is_active" json:"is_active"`
	CreatedAt       time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt       time.Time     `bson:"updated_at" json:"updated_at"`
}
