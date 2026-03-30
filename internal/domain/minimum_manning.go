package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MinimumManningStatus string

const (
	MinimumManningStatusActive      MinimumManningStatus = "active"
	MinimumManningStatusDeactivated MinimumManningStatus = "deactivated"
)

type MinimumManningEvent struct {
	ID                   bson.ObjectID        `bson:"_id,omitempty" json:"id"`
	VesselID             bson.ObjectID        `bson:"vessel_id" json:"vessel_id"`
	ActivatedByUserID    bson.ObjectID        `bson:"activated_by_user_id" json:"activated_by_user_id"`
	DeactivatedByUserID  *bson.ObjectID       `bson:"deactivated_by_user_id,omitempty" json:"deactivated_by_user_id,omitempty"`
	Reason               string               `bson:"reason,omitempty" json:"reason,omitempty"`
	ReducedPOBCap        int                  `bson:"reduced_pob_cap" json:"reduced_pob_cap"`
	Status               MinimumManningStatus `bson:"status" json:"status"`
	ActivatedAt          time.Time            `bson:"activated_at" json:"activated_at"`
	DeactivatedAt        *time.Time           `bson:"deactivated_at,omitempty" json:"deactivated_at,omitempty"`
	AffectedActivityIDs  []bson.ObjectID      `bson:"affected_activity_ids,omitempty" json:"affected_activity_ids,omitempty"`
	AffectedPersonnelIDs []bson.ObjectID      `bson:"affected_personnel_ids,omitempty" json:"affected_personnel_ids,omitempty"`
	CreatedAt            time.Time            `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time            `bson:"updated_at" json:"updated_at"`
}
