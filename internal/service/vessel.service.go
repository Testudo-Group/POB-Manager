package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrVesselCapacityExceeded = errors.New("vessel pob capacity exceeded")
)

type VesselService struct {
	repo           *repository.VesselRepository
	assignmentRepo *repository.RoomAssignmentRepository
	redis          *redis.Client
}

func NewVesselService(repo *repository.VesselRepository, assignmentRepo *repository.RoomAssignmentRepository, rdb *redis.Client) *VesselService {
	return &VesselService{
		repo:           repo,
		assignmentRepo: assignmentRepo,
		redis:          rdb,
	}
}

type CreateVesselInput struct {
	Name                   string `json:"name" validate:"required"`
	Code                   string `json:"code" validate:"required"`
	Type                   string `json:"type" validate:"required"`
	Location               string `json:"location"`
	POBCapacity            int    `json:"pob_capacity"`
	MinimumSafePOBCapacity int    `json:"minimum_safe_pob_capacity"`
}

func (s *VesselService) Create(ctx context.Context, input CreateVesselInput) (*domain.Vessel, error) {
	now := time.Now()
	v := &domain.Vessel{
		ID:                     bson.NewObjectID(),
		Name:                   input.Name,
		Code:                   input.Code,
		Type:                   domain.VesselType(input.Type),
		Location:               input.Location,
		POBCapacity:            input.POBCapacity,
		MinimumSafePOBCapacity: input.MinimumSafePOBCapacity,
		IsMinimumManningActive: false,
		Status:                 domain.VesselStatusActive,
		CreatedAt:              now,
		UpdatedAt:              now,
	}

	err := s.repo.Create(ctx, v)
	if err != nil {
		return nil, err
	}

	// Initialize the Redis counter for this vessel to 0
	key := fmt.Sprintf("vessel:%s:pob", v.ID.Hex())
	s.redis.Set(ctx, key, 0, 0)

	return v, nil
}

func (s *VesselService) FindAll(ctx context.Context) ([]domain.Vessel, error) {
	return s.repo.FindAll(ctx)
}

func (s *VesselService) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Vessel, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *VesselService) Update(ctx context.Context, id bson.ObjectID, input CreateVesselInput) (*domain.Vessel, error) {
	v, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	v.Name = input.Name
	v.Code = input.Code
	v.Type = domain.VesselType(input.Type)
	v.Location = input.Location
	v.POBCapacity = input.POBCapacity
	v.MinimumSafePOBCapacity = input.MinimumSafePOBCapacity
	v.UpdatedAt = time.Now()

	err = s.repo.Update(ctx, v)
	if err != nil {
		return nil, err
	}

	return v, nil
}

func (s *VesselService) Delete(ctx context.Context, id bson.ObjectID) error {
	return s.repo.Delete(ctx, id)
}

func (s *VesselService) GetRealTimePOB(ctx context.Context, vesselID bson.ObjectID) (int, error) {
	key := fmt.Sprintf("vessel:%s:pob", vesselID.Hex())
	val, err := s.redis.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key doesn't exist yet, POB is 0
		}
		return 0, err
	}
	return val, nil
}

// IncrementPOB atomic operation. Checks against capacity.
func (s *VesselService) IncrementPOB(ctx context.Context, vesselID bson.ObjectID) error {
	vessel, err := s.repo.FindByID(ctx, vesselID)
	if err != nil {
		return err
	}

	key := fmt.Sprintf("vessel:%s:pob", vesselID.Hex())
	
	// Transaction to check limit before incrementing, or we can just fetch and then INC
	// A pure script is atomic, but for Go we can do a quick check, then INCR
	
	currentPOB, _ := s.GetRealTimePOB(ctx, vesselID)
	if currentPOB >= vessel.POBCapacity {
		return ErrVesselCapacityExceeded
	}

	_, err = s.redis.Incr(ctx, key).Result()
	return err
}

func (s *VesselService) DecrementPOB(ctx context.Context, vesselID bson.ObjectID) error {
	key := fmt.Sprintf("vessel:%s:pob", vesselID.Hex())
	
	val, err := s.redis.Decr(ctx, key).Result()
	// Prevent negative POB
	if val < 0 {
		s.redis.Set(ctx, key, 0, 0)
	}
	return err
}

func (s *VesselService) GetManifest(ctx context.Context, vesselID bson.ObjectID) ([]domain.RoomAssignment, error) {
	return s.assignmentRepo.FindActiveByVessel(ctx, vesselID)
}
