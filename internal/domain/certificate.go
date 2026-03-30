package domain

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type CertificateStatus string

const (
	CertificateStatusValid    CertificateStatus = "valid"
	CertificateStatusExpiring CertificateStatus = "expiring"
	CertificateStatusExpired  CertificateStatus = "expired"
	CertificateStatusRevoked  CertificateStatus = "revoked"
)

type Certificate struct {
	ID                bson.ObjectID     `bson:"_id,omitempty" json:"id"`
	PersonnelID       bson.ObjectID     `bson:"personnel_id" json:"personnel_id"`
	CertificateType   string            `bson:"certificate_type" json:"certificate_type"`
	CertificateNumber string            `bson:"certificate_number" json:"certificate_number"`
	IssuedBy          string            `bson:"issued_by" json:"issued_by"`
	IssuedAt          time.Time         `bson:"issued_at" json:"issued_at"`
	ExpiresAt         time.Time         `bson:"expires_at" json:"expires_at"`
	DocumentURL       string            `bson:"document_url" json:"document_url"`
	Status            CertificateStatus `bson:"status" json:"status"`
	UploadedByUserID  *bson.ObjectID    `bson:"uploaded_by_user_id,omitempty" json:"uploaded_by_user_id,omitempty"`
	VerifiedAt        *time.Time        `bson:"verified_at,omitempty" json:"verified_at,omitempty"`
	Notes             string            `bson:"notes,omitempty" json:"notes,omitempty"`
	CreatedAt         time.Time         `bson:"created_at" json:"created_at"`
	UpdatedAt         time.Time         `bson:"updated_at" json:"updated_at"`
}
