package service

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidUserID     = errors.New("invalid user ID format")
	ErrInvalidUserRole   = errors.New("invalid role")
	ErrUserAlreadyExists = errors.New("email already exists")
)

type CreateUserReq struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=8"`
	Role        string `json:"role" binding:"required"`
	VesselID    string `json:"vessel_id"`
}

type UpdateUserReq struct {
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Email       string `json:"email" binding:"required,email"`
	VesselID    string `json:"vessel_id"`
}

type UpdateUserRoleReq struct {
	Role string `json:"role" binding:"required"`
}

type UserService struct {
	userRepo *repository.UserRepository
}

func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, organizationID string, req CreateUserReq) (*domain.User, error) {
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	role := domain.UserRole(strings.TrimSpace(req.Role))
	if !isValidUserRole(role) {
		return nil, ErrInvalidUserRole
	}

	email := strings.ToLower(strings.TrimSpace(req.Email))
	_, err = s.userRepo.FindByEmail(ctx, email)
	if err == nil {
		return nil, ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return nil, err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	var vesselID *bson.ObjectID
	if strings.TrimSpace(req.VesselID) != "" {
		parsedVesselID, err := bson.ObjectIDFromHex(strings.TrimSpace(req.VesselID))
		if err != nil {
			return nil, ErrInvalidUserID
		}
		vesselID = &parsedVesselID
	}

	now := time.Now().UTC()
	user := &domain.User{
		ID:             bson.NewObjectID(),
		OrganizationID: orgID,
		FirstName:      strings.TrimSpace(req.FirstName),
		LastName:       strings.TrimSpace(req.LastName),
		PhoneNumber:    strings.TrimSpace(req.PhoneNumber),
		Email:          email,
		PasswordHash:   string(passwordHash),
		Role:           role,
		VesselID:       vesselID,
		IsActive:       true,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, organizationID string) ([]*domain.User, error) {
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	return s.userRepo.FindAllByOrganization(ctx, orgID)
}

func (s *UserService) GetUserByID(ctx context.Context, organizationID, idStr string) (*domain.User, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	return s.userRepo.FindByIDAndOrganization(ctx, id, orgID)
}

func (s *UserService) UpdateUser(ctx context.Context, organizationID, idStr string, req UpdateUserReq) (*domain.User, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	var vesselID *bson.ObjectID
	if strings.TrimSpace(req.VesselID) != "" {
		parsedVesselID, err := bson.ObjectIDFromHex(strings.TrimSpace(req.VesselID))
		if err != nil {
			return nil, ErrInvalidUserID
		}
		vesselID = &parsedVesselID
	}

	if err := s.userRepo.UpdateUser(
		ctx,
		id,
		orgID,
		strings.TrimSpace(req.FirstName),
		strings.TrimSpace(req.LastName),
		strings.TrimSpace(req.PhoneNumber),
		strings.ToLower(strings.TrimSpace(req.Email)),
		vesselID,
	); err != nil {
		return nil, err
	}

	return s.userRepo.FindByIDAndOrganization(ctx, id, orgID)
}

func (s *UserService) DeactivateUser(ctx context.Context, organizationID, idStr string) error {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return ErrInvalidUserID
	}
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return ErrInvalidUserID
	}
	return s.userRepo.ToggleUserStatus(ctx, id, orgID, false)
}

func (s *UserService) UpdateRole(ctx context.Context, organizationID, idStr string, roleStr string) (*domain.User, error) {
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		return nil, ErrInvalidUserID
	}
	orgID, err := bson.ObjectIDFromHex(organizationID)
	if err != nil {
		return nil, ErrInvalidUserID
	}

	role := domain.UserRole(strings.TrimSpace(roleStr))
	if !isValidUserRole(role) {
		return nil, ErrInvalidUserRole
	}

	if err := s.userRepo.UpdateRole(ctx, id, orgID, role); err != nil {
		return nil, err
	}

	return s.userRepo.FindByIDAndOrganization(ctx, id, orgID)
}

func isValidUserRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleActivityOwner, domain.RolePlanner, domain.RoleSafetyAdmin, domain.RoleOIM, domain.RolePersonnel, domain.RoleSystemAdmin:
		return true
	default:
		return false
	}
}
