package service

import (
	"context"
	"errors"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrRoomCapacityExceeded = errors.New("room capacity exceeded")
	ErrVesselCapacityBlocked = errors.New("cannot assign room: vessel pob capacity exceeded")
)

type RoomService struct {
	roomRepo       *repository.RoomRepository
	assignmentRepo *repository.RoomAssignmentRepository
	vesselSvc      *VesselService
}

func NewRoomService(r *repository.RoomRepository, a *repository.RoomAssignmentRepository, v *VesselService) *RoomService {
	return &RoomService{
		roomRepo:       r,
		assignmentRepo: a,
		vesselSvc:      v,
	}
}

type CreateRoomInput struct {
	VesselID bson.ObjectID `json:"vessel_id" validate:"required"`
	Name     string        `json:"name" validate:"required"`
	Code     string        `json:"code" validate:"required"`
	Deck     string        `json:"deck"`
	Category string        `json:"category"` // "Dedicated" vs "Transient"
	Capacity int           `json:"capacity" validate:"min=1"`
}

func (s *RoomService) Create(ctx context.Context, input CreateRoomInput) (*domain.Room, error) {
	now := time.Now()
	r := &domain.Room{
		ID:        bson.NewObjectID(),
		VesselID:  input.VesselID,
		Name:      input.Name,
		Code:      input.Code,
		Deck:      input.Deck,
		Category:  input.Category,
		Capacity:  input.Capacity,
		Status:    domain.RoomStatusAvailable,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err := s.roomRepo.Create(ctx, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

type AssignRoomInput struct {
	VesselID       bson.ObjectID  `json:"vessel_id" validate:"required"`
	RoomID         bson.ObjectID  `json:"room_id" validate:"required"`
	PersonnelID    bson.ObjectID  `json:"personnel_id" validate:"required"`
	Type           string         `json:"type" validate:"required"` // "Core" or "Flexible"
	OffshoreRoleID *bson.ObjectID `json:"offshore_role_id"`         // Required if Type == "Core"
	ActivityID     *bson.ObjectID `json:"activity_id"`              // Required if Type == "Flexible"
	StartsAt       time.Time      `json:"starts_at" validate:"required"`
	EndsAt         *time.Time     `json:"ends_at"`                  // Required if Type == "Flexible"
}

func (s *RoomService) Assign(ctx context.Context, input AssignRoomInput) (*domain.RoomAssignment, error) {
	// 1. Verify Room Exists
	room, err := s.roomRepo.FindByID(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}

	// 2. Check Room Capacity
	activeAssignments, err := s.assignmentRepo.FindActiveByRoom(ctx, input.RoomID)
	if err != nil {
		return nil, err
	}
	
	if len(activeAssignments) >= room.Capacity {
		return nil, ErrRoomCapacityExceeded
	}

	// 3. ATTEMPT to Increment POB (This throws HTTP 400 equivalent if vessel is full)
	err = s.vesselSvc.IncrementPOB(ctx, input.VesselID)
	if err != nil {
		if errors.Is(err, ErrVesselCapacityExceeded) {
			return nil, ErrVesselCapacityBlocked
		}
		return nil, err
	}

	// 4. Create the assignment since POB check passed
	now := time.Now()
	assignment := &domain.RoomAssignment{
		ID:             bson.NewObjectID(),
		VesselID:       input.VesselID,
		RoomID:         input.RoomID,
		PersonnelID:    input.PersonnelID,
		OffshoreRoleID: input.OffshoreRoleID,
		ActivityID:     input.ActivityID,
		StartsAt:       input.StartsAt,
		EndsAt:         input.EndsAt,
		Status:         domain.RoomAssignmentStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	err = s.assignmentRepo.Create(ctx, assignment)
	if err != nil {
		// Rollback POB increment if db insert failed
		_ = s.vesselSvc.DecrementPOB(ctx, input.VesselID)
		return nil, err
	}

	return assignment, nil
}
