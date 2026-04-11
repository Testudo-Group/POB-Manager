package service

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidDates = errors.New("certificate expiry date must be after issue date")
var ErrInvalidCertificateType = errors.New("certificate type is not valid")

type CertificateService struct {
	repo     *repository.CertificateRepository
	typeRepo *repository.CertificateTypeRepository
}

func NewCertificateService(repo *repository.CertificateRepository, typeRepo *repository.CertificateTypeRepository) *CertificateService {
	return &CertificateService{
		repo:     repo,
		typeRepo: typeRepo,
	}
}

type CreateCertificateInput struct {
	PersonnelID       bson.ObjectID `json:"personnel_id" validate:"required"`
	CertificateType   string        `json:"certificate_type" validate:"required"`
	CertificateNumber string        `json:"certificate_number" validate:"required"`
	IssuedBy          string        `json:"issued_by"`
	IssuedAt          time.Time     `json:"issued_at"`
	ExpiresAt         time.Time     `json:"expires_at"`
	DocumentURL       string        `json:"document_url"`
}

func (s *CertificateService) Create(ctx context.Context, input CreateCertificateInput) (*domain.Certificate, error) {
	if input.ExpiresAt.Before(input.IssuedAt) {
		return nil, ErrInvalidDates
	}

	_, err := s.typeRepo.FindByCode(ctx, input.CertificateType)
	if err != nil {
		return nil, ErrInvalidCertificateType
	}

	now := time.Now()
	
	status := domain.CertificateStatusValid
	if now.After(input.ExpiresAt) {
		status = domain.CertificateStatusExpired
	} else if input.ExpiresAt.Sub(now) < 30*24*time.Hour { // Expiring in < 30 days
		status = domain.CertificateStatusExpiring
	}

	cert := &domain.Certificate{
		ID:                bson.NewObjectID(),
		PersonnelID:       input.PersonnelID,
		CertificateType:   input.CertificateType,
		CertificateNumber: input.CertificateNumber,
		IssuedBy:          input.IssuedBy,
		IssuedAt:          input.IssuedAt,
		ExpiresAt:         input.ExpiresAt,
		DocumentURL:       input.DocumentURL,
		Status:            status,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err = s.repo.Create(ctx, cert)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *CertificateService) FindByPersonnelID(ctx context.Context, personnelID bson.ObjectID) ([]domain.Certificate, error) {
	return s.repo.FindByPersonnelID(ctx, personnelID)
}

func (s *CertificateService) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Certificate, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *CertificateService) Update(ctx context.Context, id bson.ObjectID, input CreateCertificateInput) (*domain.Certificate, error) {
	if input.ExpiresAt.Before(input.IssuedAt) {
		return nil, ErrInvalidDates
	}

	_, err := s.typeRepo.FindByCode(ctx, input.CertificateType)
	if err != nil {
		return nil, ErrInvalidCertificateType
	}

	cert, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	
	status := domain.CertificateStatusValid
	if now.After(input.ExpiresAt) {
		status = domain.CertificateStatusExpired
	} else if input.ExpiresAt.Sub(now) < 30*24*time.Hour {
		status = domain.CertificateStatusExpiring
	}

	cert.CertificateType = input.CertificateType
	cert.CertificateNumber = input.CertificateNumber
	cert.IssuedBy = input.IssuedBy
	cert.IssuedAt = input.IssuedAt
	cert.ExpiresAt = input.ExpiresAt
	cert.DocumentURL = input.DocumentURL
	cert.Status = status
	cert.UpdatedAt = now

	err = s.repo.Update(ctx, cert)
	if err != nil {
		return nil, err
	}

	return cert, nil
}

func (s *CertificateService) Delete(ctx context.Context, id bson.ObjectID) error {
	return s.repo.Delete(ctx, id)
}
