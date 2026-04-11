package service

import (
	"context"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OffshoreRoleService struct {
	repo *repository.OffshoreRoleRepository
}

func NewOffshoreRoleService(repo *repository.OffshoreRoleRepository) *OffshoreRoleService {
	return &OffshoreRoleService{
		repo: repo,
	}
}

type CreateOffshoreRoleInput struct {
	Name                     string                  `json:"name" validate:"required"`
	Code                     string                  `json:"code" validate:"required"`
	Description              string                  `json:"description"`
	Type                     domain.OffshoreRoleType `json:"type" validate:"required"`
	RequiredCertificateTypes []bson.ObjectID         `json:"required_certificate_types"`
	RequiresRoom             bool                    `json:"requires_room"`
}

func (s *OffshoreRoleService) Create(ctx context.Context, input CreateOffshoreRoleInput) (*domain.OffshoreRole, error) {
	now := time.Now()
	role := &domain.OffshoreRole{
		ID:                       bson.NewObjectID(),
		Name:                     input.Name,
		Code:                     input.Code,
		Description:              input.Description,
		Type:                     input.Type,
		RequiresRoom:             input.RequiresRoom,
		RequiredCertificateTypes: input.RequiredCertificateTypes,
		Status:                   domain.OffshoreRoleStatusActive,
		CreatedAt:                now,
		UpdatedAt:                now,
	}

	err := s.repo.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	return role, nil
}

func (s *OffshoreRoleService) FindAll(ctx context.Context) ([]domain.OffshoreRole, error) {
	return s.repo.FindAll(ctx)
}
