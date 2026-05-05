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
	RequiredCertificateTypes []string                `json:"required_certificate_types"`
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

func (s *OffshoreRoleService) FindByID(ctx context.Context, id bson.ObjectID) (*domain.OffshoreRole, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *OffshoreRoleService) Update(ctx context.Context, id bson.ObjectID, input CreateOffshoreRoleInput) (*domain.OffshoreRole, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	role.Name = input.Name
	role.Code = input.Code
	role.Description = input.Description
	role.Type = input.Type
	role.RequiresRoom = input.RequiresRoom
	role.RequiredCertificateTypes = input.RequiredCertificateTypes
	role.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *OffshoreRoleService) AddRequiredCertificate(ctx context.Context, roleID bson.ObjectID, certTypeCode string) (*domain.OffshoreRole, error) {
	role, err := s.repo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	for _, existing := range role.RequiredCertificateTypes {
		if existing == certTypeCode {
			return role, nil
		}
	}

	role.RequiredCertificateTypes = append(role.RequiredCertificateTypes, certTypeCode)
	role.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}

func (s *OffshoreRoleService) RemoveRequiredCertificate(ctx context.Context, roleID bson.ObjectID, certTypeCode string) (*domain.OffshoreRole, error) {
	role, err := s.repo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	updated := make([]string, 0, len(role.RequiredCertificateTypes))
	for _, existing := range role.RequiredCertificateTypes {
		if existing != certTypeCode {
			updated = append(updated, existing)
		}
	}
	role.RequiredCertificateTypes = updated
	role.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, role); err != nil {
		return nil, err
	}
	return role, nil
}
