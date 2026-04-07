package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Organization struct {
	ID        bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name      string        `bson:"name" json:"name"`
	Phone     string        `bson:"phone" json:"phone"`
	Address   string        `bson:"address" json:"address"`
	IsActive  bool          `bson:"is_active" json:"is_active"`
	CreatedAt time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time     `bson:"updated_at" json:"updated_at"`
}
