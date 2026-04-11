package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CertificateType struct {
	ID                   bson.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                 string        `bson:"name" json:"name"`
	Code                 string        `bson:"code" json:"code"`
	Description          string        `bson:"description,omitempty" json:"description,omitempty"`
	ValidityPeriodMonths int           `bson:"validity_period_months" json:"validity_period_months"`
	IsActive             bool          `bson:"is_active" json:"is_active"`
	CreatedAt            time.Time     `bson:"created_at" json:"created_at"`
	UpdatedAt            time.Time     `bson:"updated_at" json:"updated_at"`
}
