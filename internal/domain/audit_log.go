package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuditLog struct {
	ID          bson.ObjectID  `bson:"_id,omitempty" json:"id"`
	ActorUserID *bson.ObjectID `bson:"actor_user_id,omitempty" json:"actor_user_id,omitempty"`
	EntityType  string         `bson:"entity_type" json:"entity_type"`
	EntityID    *bson.ObjectID `bson:"entity_id,omitempty" json:"entity_id,omitempty"`
	Action      string         `bson:"action" json:"action"`
	Metadata    bson.M         `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt   time.Time      `bson:"created_at" json:"created_at"`
}
