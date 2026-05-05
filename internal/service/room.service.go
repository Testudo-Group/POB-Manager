package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrRoomCapacityExceeded = errors.New("room capacity exceeded")
	ErrVesselCapacityBlocked = errors.New("cannot assign room: vessel pob capacity exceeded")
	ErrPersonnelNonCompliant = errors.New("cannot assign room: personnel is non-compliant")
)

type RoomService struct {
	roomRepo       *repository.RoomRepository
	assignmentRepo *repository.RoomAssignmentRepository
	vesselSvc      *VesselService
	compSvc        *ComplianceService
}

func NewRoomService(r *repository.RoomRepository, a *repository.RoomAssignmentRepository, v *VesselService, c *ComplianceService) *RoomService {
	return &RoomService{
		roomRepo:       r,
		assignmentRepo: a,
		vesselSvc:      v,
		compSvc:        c,
	}
}

type CreateRoomInput struct {
	VesselID    bson.ObjectID `json:"vessel_id"`
	RoomName    string        `json:"room_name"`
	RoomCategory string       `json:"room_category"`
	LocationDeck string       `json:"location_deck"`
	Capacity    int           `json:"capacity"`
	Description string        `json:"description"`
}

func roomCode(name string, id bson.ObjectID) string {
	slug := strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), " ", "-"))
	if slug == "" {
		slug = "ROOM"
	}
	return fmt.Sprintf("%s-%s", slug, id.Hex()[18:])
}

func (s *RoomService) Create(ctx context.Context, input CreateRoomInput) (*domain.Room, error) {
	if strings.TrimSpace(input.RoomName) == "" {
		return nil, errors.New("room name is required")
	}
	if input.Capacity < 1 {
		return nil, errors.New("capacity must be at least 1")
	}
	now := time.Now()
	id := bson.NewObjectID()
	r := &domain.Room{
		ID:          id,
		VesselID:    input.VesselID,
		Name:        input.RoomName,
		Code:        roomCode(input.RoomName, id),
		Deck:        input.LocationDeck,
		Category:    input.RoomCategory,
		Capacity:    input.Capacity,
		Description: input.Description,
		Status:      domain.RoomStatusAvailable,
		CreatedAt:   now,
		UpdatedAt:   now,
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

	// 3. CHECK Compliance (Travel Blocking)
	comp, err := s.compSvc.CheckCompliance(ctx, input.PersonnelID)
	if err != nil {
		return nil, err
	}
	if comp.Status != "Compliant" {
		return nil, ErrPersonnelNonCompliant
	}

	// 4. ATTEMPT to Increment POB
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

func (s *RoomService) ListByVessel(ctx context.Context, vesselID bson.ObjectID) ([]domain.Room, error) {
	rooms, err := s.roomRepo.FindByVessel(ctx, vesselID)
	if err != nil {
		return nil, err
	}
	for i, room := range rooms {
		active, err := s.assignmentRepo.FindActiveByRoom(ctx, room.ID)
		if err == nil {
			rooms[i].CurrentOccupancy = len(active)
		}
	}
	return rooms, nil
}

func (s *RoomService) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Room, error) {
	return s.roomRepo.FindByID(ctx, id)
}

func (s *RoomService) Update(ctx context.Context, id bson.ObjectID, input CreateRoomInput) (*domain.Room, error) {
	r, err := s.roomRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(input.RoomName) != "" {
		r.Name = input.RoomName
	}
	if input.LocationDeck != "" {
		r.Deck = input.LocationDeck
	}
	if input.RoomCategory != "" {
		r.Category = input.RoomCategory
	}
	if input.Capacity > 0 {
		r.Capacity = input.Capacity
	}
	if input.Description != "" {
		r.Description = input.Description
	}
	r.UpdatedAt = time.Now()

	err = s.roomRepo.Update(ctx, r)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (s *RoomService) Delete(ctx context.Context, id bson.ObjectID) error {
	// For simplicity, we assume soft delete or direct delete if no active assignments
	return s.roomRepo.Delete(ctx, id)
}

func (s *RoomService) GetOccupants(ctx context.Context, roomID bson.ObjectID) ([]domain.RoomAssignment, error) {
	return s.assignmentRepo.FindActiveByRoom(ctx, roomID)
}

func (s *RoomService) ListByDeck(ctx context.Context, vesselID bson.ObjectID) (map[string][]domain.Room, error) {
	rooms, err := s.ListByVessel(ctx, vesselID)
	if err != nil {
		return nil, err
	}

	grouped := make(map[string][]domain.Room)
	for _, room := range rooms {
		deck := room.Deck
		if deck == "" {
			deck = "Unassigned"
		}
		grouped[deck] = append(grouped[deck], room)
	}
	return grouped, nil
}
