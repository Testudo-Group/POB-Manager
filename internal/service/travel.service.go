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
	ErrTransportNotFound        = errors.New("transport not found")
	ErrTravelScheduleNotFound   = errors.New("travel schedule not found")
	ErrSeatCapacityExceeded     = errors.New("seat capacity exceeded")
	ErrTripUtilizationLow       = errors.New("trip utilization is below threshold")
	ErrPersonnelAlreadyAssigned = errors.New("personnel already assigned to this trip")
)

type TravelService struct {
	transportRepo   *repository.TransportRepository
	scheduleRepo    *repository.TravelScheduleRepository
	assignmentRepo  *repository.TravelAssignmentRepository
	activityRepo    *repository.ActivityRepository
	personnelRepo   *repository.PersonnelRepository
	complianceSvc   *ComplianceService
}

func NewTravelService(
	transportRepo *repository.TransportRepository,
	scheduleRepo *repository.TravelScheduleRepository,
	assignmentRepo *repository.TravelAssignmentRepository,
	activityRepo *repository.ActivityRepository,
	personnelRepo *repository.PersonnelRepository,
	complianceSvc *ComplianceService,
) *TravelService {
	return &TravelService{
		transportRepo:  transportRepo,
		scheduleRepo:   scheduleRepo,
		assignmentRepo: assignmentRepo,
		activityRepo:   activityRepo,
		personnelRepo:  personnelRepo,
		complianceSvc:  complianceSvc,
	}
}

// Transport Configuration
type CreateTransportInput struct {
	Name                 string                   `json:"name" validate:"required"`
	Type                 domain.TransportType     `json:"type" validate:"required"`
	Capacity             int                      `json:"capacity" validate:"required,min=1"`
	CostModel            domain.TransportCostModel `json:"cost_model" validate:"required"`
	CostAmount           float64                  `json:"cost_amount" validate:"required,min=0"`
	DepartureDays        []string                 `json:"departure_days" validate:"required"`
	MobilizationLocation string                   `json:"mobilization_location" validate:"required"`
}

func (s *TravelService) CreateTransport(ctx context.Context, input CreateTransportInput) (*domain.Transport, error) {
	now := time.Now()
	transport := &domain.Transport{
		ID:                   bson.NewObjectID(),
		Name:                 input.Name,
		Type:                 input.Type,
		Capacity:             input.Capacity,
		CostModel:            input.CostModel,
		CostAmount:           input.CostAmount,
		DepartureDays:        input.DepartureDays,
		MobilizationLocation: input.MobilizationLocation,
		IsActive:             true,
		CreatedAt:            now,
		UpdatedAt:            now,
	}
	err := s.transportRepo.Create(ctx, transport)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

func (s *TravelService) ListTransports(ctx context.Context, activeOnly bool) ([]domain.Transport, error) {
	var isActive *bool
	if activeOnly {
		isActive = &activeOnly
	}
	return s.transportRepo.FindAll(ctx, isActive)
}

func (s *TravelService) GetTransport(ctx context.Context, id bson.ObjectID) (*domain.Transport, error) {
	return s.transportRepo.FindByID(ctx, id)
}

func (s *TravelService) UpdateTransport(ctx context.Context, id bson.ObjectID, input CreateTransportInput) (*domain.Transport, error) {
	transport, err := s.transportRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	transport.Name = input.Name
	transport.Type = input.Type
	transport.Capacity = input.Capacity
	transport.CostModel = input.CostModel
	transport.CostAmount = input.CostAmount
	transport.DepartureDays = input.DepartureDays
	transport.MobilizationLocation = input.MobilizationLocation
	transport.UpdatedAt = time.Now()

	err = s.transportRepo.Update(ctx, transport)
	if err != nil {
		return nil, err
	}
	return transport, nil
}

func (s *TravelService) DeleteTransport(ctx context.Context, id bson.ObjectID) error {
	return s.transportRepo.Delete(ctx, id)
}

// Travel Schedule
type CreateTravelScheduleInput struct {
	TransportID   bson.ObjectID           `json:"transport_id" validate:"required"`
	VesselID      *bson.ObjectID          `json:"vessel_id"`
	ActivityID    *bson.ObjectID          `json:"activity_id"`
	Direction     domain.TravelDirection  `json:"direction" validate:"required"`
	DepartureAt   time.Time               `json:"departure_at" validate:"required"`
}

func (s *TravelService) CreateTravelSchedule(ctx context.Context, input CreateTravelScheduleInput) (*domain.TravelSchedule, error) {
	transport, err := s.transportRepo.FindByID(ctx, input.TransportID)
	if err != nil {
		return nil, ErrTransportNotFound
	}

	now := time.Now()
	schedule := &domain.TravelSchedule{
		ID:           bson.NewObjectID(),
		TransportID:  input.TransportID,
		VesselID:     input.VesselID,
		ActivityID:   input.ActivityID,
		Direction:    input.Direction,
		DepartureAt:  input.DepartureAt,
		SeatCapacity: transport.Capacity,
		ReservedSeats: 0,
		Status:       domain.TravelScheduleStatusPlanned,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = s.scheduleRepo.Create(ctx, schedule)
	if err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *TravelService) GetTravelSchedule(ctx context.Context, id bson.ObjectID) (*domain.TravelSchedule, error) {
	return s.scheduleRepo.FindByID(ctx, id)
}

func (s *TravelService) ListUpcomingSchedules(ctx context.Context, limit int) ([]domain.TravelSchedule, error) {
	return s.scheduleRepo.FindUpcoming(ctx, limit)
}

// Auto-match activities to transport
func (s *TravelService) MatchActivitiesToTransport(ctx context.Context, transportID bson.ObjectID, startDate, endDate time.Time) ([]domain.Activity, error) {
	// Find transport to check departure days
	transport, err := s.transportRepo.FindByID(ctx, transportID)
	if err != nil {
		return nil, err
	}

	// Find activities starting in the date range
	activities, err := s.activityRepo.FindByDateRange(ctx, bson.NilObjectID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	// Filter activities whose start date matches a departure day
	var matched []domain.Activity
	departureDayMap := make(map[string]bool)
	for _, d := range transport.DepartureDays {
		departureDayMap[d] = true
	}

	for _, a := range activities {
		if a.Status == domain.ActivityStatusApproved {
			dayName := a.StartDate.Weekday().String()
			if departureDayMap[dayName] {
				matched = append(matched, a)
			}
		}
	}
	return matched, nil
}

// Personnel Assignment
type AssignPersonnelToTripInput struct {
	TravelScheduleID bson.ObjectID   `json:"travel_schedule_id" validate:"required"`
	PersonnelIDs     []bson.ObjectID `json:"personnel_ids" validate:"required"`
}

func (s *TravelService) AssignPersonnelToTrip(ctx context.Context, input AssignPersonnelToTripInput) error {
	schedule, err := s.scheduleRepo.FindByID(ctx, input.TravelScheduleID)
	if err != nil {
		return err
	}

	// Check seat availability
	currentCount, err := s.assignmentRepo.CountBySchedule(ctx, schedule.ID)
	if err != nil {
		return err
	}
	if int(currentCount)+len(input.PersonnelIDs) > schedule.SeatCapacity {
		return ErrSeatCapacityExceeded
	}

	// Check compliance for each personnel
	for _, pID := range input.PersonnelIDs {
		comp, err := s.complianceSvc.CheckCompliance(ctx, pID)
		if err != nil {
			return err
		}
		if comp.Status != "Compliant" {
			return errors.New("personnel " + pID.Hex() + " is not compliant for travel")
		}
	}

	now := time.Now()
	var assignments []domain.TravelAssignment
	for _, pID := range input.PersonnelIDs {
		assignments = append(assignments, domain.TravelAssignment{
			ID:               bson.NewObjectID(),
			TravelScheduleID: schedule.ID,
			PersonnelID:      pID,
			Status:           domain.TravelAssignmentStatusReserved,
			AssignedAt:       now,
			CreatedAt:        now,
			UpdatedAt:        now,
		})
	}

	err = s.assignmentRepo.CreateMany(ctx, assignments)
	if err != nil {
		return err
	}

	// Update reserved seats count
	schedule.ReservedSeats += len(input.PersonnelIDs)
	return s.scheduleRepo.Update(ctx, schedule)
}

// Utilization Alerts
type UtilizationAlert struct {
	ScheduleID   bson.ObjectID `json:"schedule_id"`
	TransportName string       `json:"transport_name"`
	DepartureAt   time.Time    `json:"departure_at"`
	Capacity      int          `json:"capacity"`
	ReservedSeats int          `json:"reserved_seats"`
	Utilization   float64      `json:"utilization_percent"`
}

func (s *TravelService) CheckLowUtilization(ctx context.Context, thresholdPercent float64) ([]UtilizationAlert, error) {
	schedules, err := s.scheduleRepo.FindUpcoming(ctx, 100)
	if err != nil {
		return nil, err
	}

	var alerts []UtilizationAlert
	for _, sch := range schedules {
		if sch.Status == domain.TravelScheduleStatusPlanned {
			utilization := float64(sch.ReservedSeats) / float64(sch.SeatCapacity) * 100
			if utilization < thresholdPercent {
				transport, _ := s.transportRepo.FindByID(ctx, sch.TransportID)
				transportName := "Unknown"
				if transport != nil {
					transportName = transport.Name
				}
				alerts = append(alerts, UtilizationAlert{
					ScheduleID:    sch.ID,
					TransportName: transportName,
					DepartureAt:   sch.DepartureAt,
					Capacity:      sch.SeatCapacity,
					ReservedSeats: sch.ReservedSeats,
					Utilization:   utilization,
				})
			}
		}
	}
	return alerts, nil
}

// Trip Consolidation Suggestions
func (s *TravelService) SuggestTripConsolidation(ctx context.Context, transportID bson.ObjectID, date time.Time) ([]domain.TravelSchedule, error) {
	// Find all schedules for the same transport within 2 days
	start := date.AddDate(0, 0, -2)
	end := date.AddDate(0, 0, 2)
	schedules, err := s.scheduleRepo.FindByTransportAndDateRange(ctx, transportID, start, end)
	if err != nil {
		return nil, err
	}

	// Return schedules that could be consolidated (low utilization)
	var candidates []domain.TravelSchedule
	for _, sch := range schedules {
		if sch.Status == domain.TravelScheduleStatusPlanned {
			utilization := float64(sch.ReservedSeats) / float64(sch.SeatCapacity)
			if utilization < 0.6 {
				candidates = append(candidates, sch)
			}
		}
	}
	return candidates, nil
}

// Get personnel travel schedule
func (s *TravelService) GetPersonnelTravels(ctx context.Context, personnelID bson.ObjectID) ([]domain.TravelAssignment, error) {
	return s.assignmentRepo.FindByPersonnel(ctx, personnelID)
}

func (s *TravelService) GetTripAssignments(ctx context.Context, scheduleID bson.ObjectID) ([]domain.TravelAssignment, error) {
	return s.assignmentRepo.FindBySchedule(ctx, scheduleID)
}