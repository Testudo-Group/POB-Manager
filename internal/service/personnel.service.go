package service

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var ErrInvalidOffshoreRole = errors.New("one or more offshore roles are invalid")

type PersonnelService struct {
	repo     *repository.PersonnelRepository
	roleRepo *repository.OffshoreRoleRepository
}

func NewPersonnelService(repo *repository.PersonnelRepository, roleRepo *repository.OffshoreRoleRepository) *PersonnelService {
	return &PersonnelService{
		repo:     repo,
		roleRepo: roleRepo,
	}
}

type CreatePersonnelInput struct {
	EmployeeNumber    string            `json:"employee_number" validate:"required"`
	FirstName         string            `json:"first_name" validate:"required"`
	LastName          string            `json:"last_name" validate:"required"`
	Email             string            `json:"email" validate:"required,email"`
	PhoneNumber       string            `json:"phone_number"`
	Nationality       string            `json:"nationality"`
	Company           string            `json:"company"`
	PrimaryDiscipline string            `json:"primary_discipline"`
	OffshoreRoleIDs   []bson.ObjectID   `json:"offshore_role_ids"`
}

func (s *PersonnelService) Create(ctx context.Context, input CreatePersonnelInput) (*domain.Personnel, error) {
	for _, roleID := range input.OffshoreRoleIDs {
		_, err := s.roleRepo.FindByID(ctx, roleID)
		if err != nil {
			return nil, ErrInvalidOffshoreRole
		}
	}

	now := time.Now()
	p := &domain.Personnel{
		ID:                bson.NewObjectID(),
		EmployeeNumber:    input.EmployeeNumber,
		FirstName:         input.FirstName,
		LastName:          input.LastName,
		Email:             input.Email,
		PhoneNumber:       input.PhoneNumber,
		Nationality:       input.Nationality,
		Company:           input.Company,
		PrimaryDiscipline: input.PrimaryDiscipline,
		OffshoreRoleIDs:   input.OffshoreRoleIDs,
		CurrentStatus:     domain.PersonnelStatusAvailable,
		IsActive:          true,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	err := s.repo.Create(ctx, p)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (s *PersonnelService) FindAll(ctx context.Context) ([]domain.Personnel, error) {
	return s.repo.FindAll(ctx)
}

func (s *PersonnelService) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Personnel, error) {
	return s.repo.FindByID(ctx, id)
}
