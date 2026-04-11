package service

import (
	"context"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CertificateTypeService struct {
	repo *repository.CertificateTypeRepository
}

func NewCertificateTypeService(repo *repository.CertificateTypeRepository) *CertificateTypeService {
	return &CertificateTypeService{
		repo: repo,
	}
}

type CreateCertificateTypeInput struct {
	Name                 string `json:"name" validate:"required"`
	Code                 string `json:"code" validate:"required"`
	Description          string `json:"description"`
	ValidityPeriodMonths int    `json:"validity_period_months"`
}

func (s *CertificateTypeService) Create(ctx context.Context, input CreateCertificateTypeInput) (*domain.CertificateType, error) {
	now := time.Now()
	ct := &domain.CertificateType{
		ID:                   bson.NewObjectID(),
		Name:                 input.Name,
		Code:                 input.Code,
		Description:          input.Description,
		ValidityPeriodMonths: input.ValidityPeriodMonths,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err := s.repo.Create(ctx, ct)
	if err != nil {
		return nil, err
	}

	return ct, nil
}

func (s *CertificateTypeService) FindAll(ctx context.Context) ([]domain.CertificateType, error) {
	return s.repo.FindAll(ctx)
}
