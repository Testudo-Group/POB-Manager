package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActivityRequirement struct {
	ID             bson.ObjectID `bson:"_id,omitempty" json:"id"`
	ActivityID     bson.ObjectID `bson:"activity_id" json:"activity_id"`
	OffshoreRoleID bson.ObjectID `bson:"offshore_role_id" json:"offshore_role_id"`
	RequiredCount  int           `bson:"required_count" json:"required_count"`
	AssignedCount  int           `bson:"assigned_count" json:"assigned_count"`
	CreatedAt      time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `bson:"updated_at" json:"updated_at"`
}
