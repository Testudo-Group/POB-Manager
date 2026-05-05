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
	ErrInvalidTravelRoute       = errors.New("origin and destination vessels must be different")
)

type TravelService struct {
	transportRepo  *repository.TransportRepository
	scheduleRepo   *repository.TravelScheduleRepository
	assignmentRepo *repository.TravelAssignmentRepository
	activityRepo   *repository.ActivityRepository
	personnelRepo  *repository.PersonnelRepository
	vesselRepo     *repository.VesselRepository
	complianceSvc  *ComplianceService
}

func NewTravelService(
	transportRepo *repository.TransportRepository,
	scheduleRepo *repository.TravelScheduleRepository,
	assignmentRepo *repository.TravelAssignmentRepository,
	activityRepo *repository.ActivityRepository,
	personnelRepo *repository.PersonnelRepository,
	vesselRepo *repository.VesselRepository,
	complianceSvc *ComplianceService,
) *TravelService {
	return &TravelService{
		transportRepo:  transportRepo,
		scheduleRepo:   scheduleRepo,
		assignmentRepo: assignmentRepo,
		activityRepo:   activityRepo,
		personnelRepo:  personnelRepo,
		vesselRepo:     vesselRepo,
		complianceSvc:  complianceSvc,
	}
}

type VesselSummary struct {
	ID       bson.ObjectID       `json:"id"`
	Name     string              `json:"name"`
	Code     string              `json:"code"`
	Type     domain.VesselType   `json:"type"`
	Location string              `json:"location"`
	Status   domain.VesselStatus `json:"status"`
}

type TransportResponse struct {
	ID                    bson.ObjectID             `json:"id"`
	Name                  string                    `json:"name"`
	Type                  domain.TransportType      `json:"type"`
	Capacity              int                       `json:"capacity"`
	CostModel             domain.TransportCostModel `json:"cost_model"`
	CostAmount            float64                   `json:"cost_amount"`
	DepartureDays         []string                  `json:"departure_days"`
	MobilizationLocation  string                    `json:"mobilization_location"`
	OriginVesselID        *bson.ObjectID            `json:"origin_vessel_id,omitempty"`
	DestinationVesselID   *bson.ObjectID            `json:"destination_vessel_id,omitempty"`
	OriginVessel          *VesselSummary            `json:"origin_vessel,omitempty"`
	DestinationVessel     *VesselSummary            `json:"destination_vessel,omitempty"`
	RouteWaypoints        []bson.ObjectID           `json:"route_waypoints,omitempty"`
	RouteWaypointVessels  []*VesselSummary          `json:"route_waypoint_vessels,omitempty"`
	TravelDurationMinutes int                       `json:"travel_duration_minutes,omitempty"`
	RouteLabel            string                    `json:"route_label,omitempty"`
	IsActive              bool                      `json:"is_active"`
	CreatedAt             time.Time                 `json:"created_at"`
	UpdatedAt             time.Time                 `json:"updated_at"`
}

type TravelScheduleResponse struct {
	ID                  bson.ObjectID               `json:"id"`
	TransportID         bson.ObjectID               `json:"transport_id"`
	TransportName       string                      `json:"transport_name,omitempty"`
	TransportType       domain.TransportType        `json:"transport_type,omitempty"`
	VesselID            *bson.ObjectID              `json:"vessel_id,omitempty"`
	OriginVesselID      *bson.ObjectID              `json:"origin_vessel_id,omitempty"`
	DestinationVesselID *bson.ObjectID              `json:"destination_vessel_id,omitempty"`
	OriginVessel        *VesselSummary              `json:"origin_vessel,omitempty"`
	DestinationVessel   *VesselSummary              `json:"destination_vessel,omitempty"`
	ActivityID          *bson.ObjectID              `json:"activity_id,omitempty"`
	Direction           domain.TravelDirection      `json:"direction"`
	DepartureAt         time.Time                   `json:"departure_at"`
	ArrivalAt           *time.Time                  `json:"arrival_at,omitempty"`
	SeatCapacity        int                         `json:"seat_capacity"`
	ReservedSeats       int                         `json:"reserved_seats"`
	AvailableSeats      int                         `json:"available_seats"`
	Status              domain.TravelScheduleStatus `json:"status"`
	RouteLabel          string                      `json:"route_label,omitempty"`
	CreatedAt           time.Time                   `json:"created_at"`
	UpdatedAt           time.Time                   `json:"updated_at"`
}

// Transport Configuration
type CreateTransportInput struct {
	Name                  string                    `json:"name" validate:"required"`
	Type                  domain.TransportType      `json:"type" validate:"required"`
	Capacity              int                       `json:"capacity" validate:"required,min=1"`
	CostModel             domain.TransportCostModel `json:"cost_model" validate:"required"`
	CostAmount            float64                   `json:"cost_amount" validate:"required,min=0"`
	DepartureDays         []string                  `json:"departure_days" validate:"required"`
	MobilizationLocation  string                    `json:"mobilization_location" validate:"required"`
	OriginVesselID        *bson.ObjectID            `json:"origin_vessel_id"`
	DestinationVesselID   *bson.ObjectID            `json:"destination_vessel_id"`
	RouteWaypoints        []bson.ObjectID           `json:"route_waypoints"`
	TravelDurationMinutes int                       `json:"travel_duration_minutes"`
}

func (s *TravelService) CreateTransport(ctx context.Context, input CreateTransportInput) (*TransportResponse, error) {
	if err := s.validateRoute(ctx, input.OriginVesselID, input.DestinationVesselID); err != nil {
		return nil, err
	}

	originID, destID := resolveWaypointEndpoints(input.RouteWaypoints, input.OriginVesselID, input.DestinationVesselID)

	now := time.Now()
	transport := &domain.Transport{
		ID:                    bson.NewObjectID(),
		Name:                  input.Name,
		Type:                  input.Type,
		Capacity:              input.Capacity,
		CostModel:             input.CostModel,
		CostAmount:            input.CostAmount,
		DepartureDays:         input.DepartureDays,
		MobilizationLocation:  input.MobilizationLocation,
		OriginVesselID:        originID,
		DestinationVesselID:   destID,
		RouteWaypoints:        input.RouteWaypoints,
		TravelDurationMinutes: input.TravelDurationMinutes,
		IsActive:              true,
		CreatedAt:             now,
		UpdatedAt:             now,
	}
	if err := s.transportRepo.Create(ctx, transport); err != nil {
		return nil, err
	}
	return s.buildTransportResponse(ctx, transport)
}

func (s *TravelService) ListTransports(ctx context.Context, activeOnly bool) ([]TransportResponse, error) {
	var isActive *bool
	if activeOnly {
		isActive = &activeOnly
	}

	transports, err := s.transportRepo.FindAll(ctx, isActive)
	if err != nil {
		return nil, err
	}

	lookup := newVesselLookup(ctx, s.vesselRepo)
	responses := make([]TransportResponse, 0, len(transports))
	for _, transport := range transports {
		resp, err := s.buildTransportResponseWithLookup(&transport, lookup)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}

func (s *TravelService) GetTransport(ctx context.Context, id bson.ObjectID) (*TransportResponse, error) {
	transport, err := s.transportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTransportNotFound) {
			return nil, ErrTransportNotFound
		}
		return nil, err
	}
	return s.buildTransportResponse(ctx, transport)
}

func (s *TravelService) UpdateTransport(ctx context.Context, id bson.ObjectID, input CreateTransportInput) (*TransportResponse, error) {
	if err := s.validateRoute(ctx, input.OriginVesselID, input.DestinationVesselID); err != nil {
		return nil, err
	}

	transport, err := s.transportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTransportNotFound) {
			return nil, ErrTransportNotFound
		}
		return nil, err
	}

	originID, destID := resolveWaypointEndpoints(input.RouteWaypoints, input.OriginVesselID, input.DestinationVesselID)

	transport.Name = input.Name
	transport.Type = input.Type
	transport.Capacity = input.Capacity
	transport.CostModel = input.CostModel
	transport.CostAmount = input.CostAmount
	transport.DepartureDays = input.DepartureDays
	transport.MobilizationLocation = input.MobilizationLocation
	transport.OriginVesselID = originID
	transport.DestinationVesselID = destID
	transport.RouteWaypoints = input.RouteWaypoints
	transport.TravelDurationMinutes = input.TravelDurationMinutes
	transport.UpdatedAt = time.Now()

	if err := s.transportRepo.Update(ctx, transport); err != nil {
		return nil, err
	}
	return s.buildTransportResponse(ctx, transport)
}

func (s *TravelService) DeleteTransport(ctx context.Context, id bson.ObjectID) error {
	return s.transportRepo.Delete(ctx, id)
}

// Travel Schedule
type CreateTravelScheduleInput struct {
	TransportID         bson.ObjectID          `json:"transport_id" validate:"required"`
	VesselID            *bson.ObjectID         `json:"vessel_id"`
	OriginVesselID      *bson.ObjectID         `json:"origin_vessel_id"`
	DestinationVesselID *bson.ObjectID         `json:"destination_vessel_id"`
	ActivityID          *bson.ObjectID         `json:"activity_id"`
	Direction           domain.TravelDirection `json:"direction" validate:"required"`
	DepartureAt         time.Time              `json:"departure_at" validate:"required"`
}

type ListTravelSchedulesInput struct {
	TransportID         *bson.ObjectID
	VesselID            *bson.ObjectID
	OriginVesselID      *bson.ObjectID
	DestinationVesselID *bson.ObjectID
	Status              *domain.TravelScheduleStatus
	UpcomingOnly        bool
	Limit               int
}

func (s *TravelService) CreateTravelSchedule(ctx context.Context, input CreateTravelScheduleInput) (*TravelScheduleResponse, error) {
	transport, err := s.transportRepo.FindByID(ctx, input.TransportID)
	if err != nil {
		return nil, ErrTransportNotFound
	}

	originVesselID, destinationVesselID, vesselID, err := s.resolveScheduleRoute(ctx, transport, input)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	schedule := &domain.TravelSchedule{
		ID:                  bson.NewObjectID(),
		TransportID:         input.TransportID,
		VesselID:            vesselID,
		OriginVesselID:      originVesselID,
		DestinationVesselID: destinationVesselID,
		ActivityID:          input.ActivityID,
		Direction:           input.Direction,
		DepartureAt:         input.DepartureAt,
		SeatCapacity:        transport.Capacity,
		ReservedSeats:       0,
		Status:              domain.TravelScheduleStatusPlanned,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	if err := s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, err
	}
	return s.buildTravelScheduleResponse(ctx, schedule, transport)
}

func (s *TravelService) GetTravelSchedule(ctx context.Context, id bson.ObjectID) (*TravelScheduleResponse, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTravelScheduleNotFound) {
			return nil, ErrTravelScheduleNotFound
		}
		return nil, err
	}
	return s.buildTravelScheduleResponse(ctx, schedule, nil)
}

func (s *TravelService) ListTravelSchedules(ctx context.Context, input ListTravelSchedulesInput) ([]TravelScheduleResponse, error) {
	limit := int64(input.Limit)
	if limit <= 0 {
		limit = 100
	}

	schedules, err := s.scheduleRepo.Find(ctx, repository.TravelScheduleFilters{
		TransportID:         input.TransportID,
		VesselID:            input.VesselID,
		OriginVesselID:      input.OriginVesselID,
		DestinationVesselID: input.DestinationVesselID,
		Status:              input.Status,
		UpcomingOnly:        input.UpcomingOnly,
		Limit:               limit,
	})
	if err != nil {
		return nil, err
	}

	lookup := newVesselLookup(ctx, s.vesselRepo)
	transportCache := make(map[string]*domain.Transport)
	responses := make([]TravelScheduleResponse, 0, len(schedules))
	for _, schedule := range schedules {
		transport, err := s.getTransportFromCache(ctx, transportCache, schedule.TransportID)
		if err != nil {
			return nil, err
		}
		resp, err := s.buildTravelScheduleResponseWithLookup(&schedule, transport, lookup)
		if err != nil {
			return nil, err
		}
		responses = append(responses, *resp)
	}
	return responses, nil
}

func (s *TravelService) ListUpcomingSchedules(ctx context.Context, limit int) ([]TravelScheduleResponse, error) {
	status := domain.TravelScheduleStatusPlanned
	return s.ListTravelSchedules(ctx, ListTravelSchedulesInput{
		Status:       &status,
		UpcomingOnly: true,
		Limit:        limit,
	})
}

// Auto-match activities to transport
func (s *TravelService) MatchActivitiesToTransport(ctx context.Context, transportID bson.ObjectID, startDate, endDate time.Time) ([]domain.Activity, error) {
	transport, err := s.transportRepo.FindByID(ctx, transportID)
	if err != nil {
		return nil, err
	}

	activities, err := s.activityRepo.FindByDateRange(ctx, bson.NilObjectID, startDate, endDate)
	if err != nil {
		return nil, err
	}

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

	currentCount, err := s.assignmentRepo.CountBySchedule(ctx, schedule.ID)
	if err != nil {
		return err
	}
	if int(currentCount)+len(input.PersonnelIDs) > schedule.SeatCapacity {
		return ErrSeatCapacityExceeded
	}

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

	if err := s.assignmentRepo.CreateMany(ctx, assignments); err != nil {
		return err
	}

	schedule.ReservedSeats += len(input.PersonnelIDs)
	return s.scheduleRepo.Update(ctx, schedule)
}

// Utilization Alerts
type UtilizationAlert struct {
	ScheduleID    bson.ObjectID `json:"schedule_id"`
	TransportName string        `json:"transport_name"`
	DepartureAt   time.Time     `json:"departure_at"`
	Capacity      int           `json:"capacity"`
	ReservedSeats int           `json:"reserved_seats"`
	Utilization   float64       `json:"utilization_percent"`
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
	start := date.AddDate(0, 0, -2)
	end := date.AddDate(0, 0, 2)
	schedules, err := s.scheduleRepo.FindByTransportAndDateRange(ctx, transportID, start, end)
	if err != nil {
		return nil, err
	}

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

func (s *TravelService) resolveScheduleRoute(ctx context.Context, transport *domain.Transport, input CreateTravelScheduleInput) (*bson.ObjectID, *bson.ObjectID, *bson.ObjectID, error) {
	originVesselID := firstNonNilObjectID(input.OriginVesselID, transport.OriginVesselID)
	destinationVesselID := firstNonNilObjectID(input.DestinationVesselID, transport.DestinationVesselID)

	if input.VesselID != nil {
		switch input.Direction {
		case domain.TravelDirectionOutbound:
			if originVesselID == nil {
				originVesselID = cloneObjectID(input.VesselID)
			}
		case domain.TravelDirectionInbound:
			if destinationVesselID == nil {
				destinationVesselID = cloneObjectID(input.VesselID)
			}
		}
	}

	if err := s.validateRoute(ctx, originVesselID, destinationVesselID); err != nil {
		return nil, nil, nil, err
	}

	vesselID := cloneObjectID(input.VesselID)
	if vesselID == nil {
		switch input.Direction {
		case domain.TravelDirectionInbound:
			vesselID = cloneObjectID(destinationVesselID)
		case domain.TravelDirectionOutbound:
			vesselID = cloneObjectID(originVesselID)
		}
	}
	if vesselID == nil {
		vesselID = firstNonNilObjectID(destinationVesselID, originVesselID)
	}

	return originVesselID, destinationVesselID, vesselID, nil
}

func (s *TravelService) validateRoute(ctx context.Context, originVesselID, destinationVesselID *bson.ObjectID) error {
	if originVesselID != nil && destinationVesselID != nil && *originVesselID == *destinationVesselID {
		return ErrInvalidTravelRoute
	}

	for _, vesselID := range []*bson.ObjectID{originVesselID, destinationVesselID} {
		if vesselID == nil {
			continue
		}
		if _, err := s.vesselRepo.FindByID(ctx, *vesselID); err != nil {
			return err
		}
	}
	return nil
}

func (s *TravelService) buildTransportResponse(ctx context.Context, transport *domain.Transport) (*TransportResponse, error) {
	return s.buildTransportResponseWithLookup(transport, newVesselLookup(ctx, s.vesselRepo))
}

func (s *TravelService) buildTransportResponseWithLookup(transport *domain.Transport, lookup *vesselLookup) (*TransportResponse, error) {
	originVessel, err := lookup.get(transport.OriginVesselID)
	if err != nil {
		return nil, err
	}
	destinationVessel, err := lookup.get(transport.DestinationVesselID)
	if err != nil {
		return nil, err
	}

	waypointVessels := make([]*VesselSummary, 0, len(transport.RouteWaypoints))
	for i := range transport.RouteWaypoints {
		id := transport.RouteWaypoints[i]
		v, err := lookup.get(&id)
		if err != nil {
			return nil, err
		}
		waypointVessels = append(waypointVessels, v)
	}

	return &TransportResponse{
		ID:                    transport.ID,
		Name:                  transport.Name,
		Type:                  transport.Type,
		Capacity:              transport.Capacity,
		CostModel:             transport.CostModel,
		CostAmount:            transport.CostAmount,
		DepartureDays:         transport.DepartureDays,
		MobilizationLocation:  transport.MobilizationLocation,
		OriginVesselID:        cloneObjectID(transport.OriginVesselID),
		DestinationVesselID:   cloneObjectID(transport.DestinationVesselID),
		OriginVessel:          originVessel,
		DestinationVessel:     destinationVessel,
		RouteWaypoints:        transport.RouteWaypoints,
		RouteWaypointVessels:  waypointVessels,
		TravelDurationMinutes: transport.TravelDurationMinutes,
		RouteLabel:            buildRouteLabel(waypointVessels, originVessel, destinationVessel),
		IsActive:              transport.IsActive,
		CreatedAt:             transport.CreatedAt,
		UpdatedAt:             transport.UpdatedAt,
	}, nil
}

func (s *TravelService) buildTravelScheduleResponse(ctx context.Context, schedule *domain.TravelSchedule, transport *domain.Transport) (*TravelScheduleResponse, error) {
	lookup := newVesselLookup(ctx, s.vesselRepo)
	if transport == nil {
		var err error
		transport, err = s.transportRepo.FindByID(ctx, schedule.TransportID)
		if err != nil && !errors.Is(err, repository.ErrTransportNotFound) {
			return nil, err
		}
	}
	return s.buildTravelScheduleResponseWithLookup(schedule, transport, lookup)
}

func (s *TravelService) buildTravelScheduleResponseWithLookup(schedule *domain.TravelSchedule, transport *domain.Transport, lookup *vesselLookup) (*TravelScheduleResponse, error) {
	originVesselID, destinationVesselID := resolvedScheduleRouteIDs(schedule, transport)

	originVessel, err := lookup.get(originVesselID)
	if err != nil {
		return nil, err
	}
	destinationVessel, err := lookup.get(destinationVesselID)
	if err != nil {
		return nil, err
	}

	resp := &TravelScheduleResponse{
		ID:                  schedule.ID,
		TransportID:         schedule.TransportID,
		VesselID:            cloneObjectID(schedule.VesselID),
		OriginVesselID:      cloneObjectID(originVesselID),
		DestinationVesselID: cloneObjectID(destinationVesselID),
		OriginVessel:        originVessel,
		DestinationVessel:   destinationVessel,
		ActivityID:          schedule.ActivityID,
		Direction:           schedule.Direction,
		DepartureAt:         schedule.DepartureAt,
		ArrivalAt:           schedule.ArrivalAt,
		SeatCapacity:        schedule.SeatCapacity,
		ReservedSeats:       schedule.ReservedSeats,
		AvailableSeats:      schedule.SeatCapacity - schedule.ReservedSeats,
		Status:              schedule.Status,
		RouteLabel:          routeLabel(originVessel, destinationVessel),
		CreatedAt:           schedule.CreatedAt,
		UpdatedAt:           schedule.UpdatedAt,
	}

	if transport != nil {
		resp.TransportName = transport.Name
		resp.TransportType = transport.Type
	}

	return resp, nil
}

func (s *TravelService) getTransportFromCache(ctx context.Context, cache map[string]*domain.Transport, id bson.ObjectID) (*domain.Transport, error) {
	if transport, ok := cache[id.Hex()]; ok {
		return transport, nil
	}

	transport, err := s.transportRepo.FindByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrTransportNotFound) {
			cache[id.Hex()] = nil
			return nil, nil
		}
		return nil, err
	}

	cache[id.Hex()] = transport
	return transport, nil
}

type vesselLookup struct {
	ctx   context.Context
	repo  *repository.VesselRepository
	cache map[string]*VesselSummary
}

func newVesselLookup(ctx context.Context, repo *repository.VesselRepository) *vesselLookup {
	return &vesselLookup{
		ctx:   ctx,
		repo:  repo,
		cache: make(map[string]*VesselSummary),
	}
}

func (l *vesselLookup) get(id *bson.ObjectID) (*VesselSummary, error) {
	if id == nil {
		return nil, nil
	}

	key := id.Hex()
	if summary, ok := l.cache[key]; ok {
		return summary, nil
	}

	vessel, err := l.repo.FindByID(l.ctx, *id)
	if err != nil {
		if errors.Is(err, repository.ErrVesselNotFound) {
			l.cache[key] = nil
			return nil, nil
		}
		return nil, err
	}

	summary := &VesselSummary{
		ID:       vessel.ID,
		Name:     vessel.Name,
		Code:     vessel.Code,
		Type:     vessel.Type,
		Location: vessel.Location,
		Status:   vessel.Status,
	}
	l.cache[key] = summary
	return summary, nil
}

func routeLabel(originVessel, destinationVessel *VesselSummary) string {
	return buildRouteLabel(nil, originVessel, destinationVessel)
}

func buildRouteLabel(waypoints []*VesselSummary, originVessel, destinationVessel *VesselSummary) string {
	if len(waypoints) > 1 {
		names := make([]string, 0, len(waypoints))
		for _, v := range waypoints {
			if v != nil {
				names = append(names, v.Name)
			}
		}
		if len(names) > 0 {
			label := names[0]
			for _, n := range names[1:] {
				label += " -> " + n
			}
			return label
		}
	}
	switch {
	case originVessel != nil && destinationVessel != nil:
		return originVessel.Name + " -> " + destinationVessel.Name
	case originVessel != nil:
		return "From " + originVessel.Name
	case destinationVessel != nil:
		return "To " + destinationVessel.Name
	default:
		return ""
	}
}

func resolveWaypointEndpoints(waypoints []bson.ObjectID, originID, destID *bson.ObjectID) (*bson.ObjectID, *bson.ObjectID) {
	if len(waypoints) >= 2 {
		first := waypoints[0]
		last := waypoints[len(waypoints)-1]
		return &first, &last
	}
	return cloneObjectID(originID), cloneObjectID(destID)
}

func cloneObjectID(id *bson.ObjectID) *bson.ObjectID {
	if id == nil {
		return nil
	}
	cloned := *id
	return &cloned
}

func firstNonNilObjectID(ids ...*bson.ObjectID) *bson.ObjectID {
	for _, id := range ids {
		if id != nil {
			return cloneObjectID(id)
		}
	}
	return nil
}

func resolvedScheduleRouteIDs(schedule *domain.TravelSchedule, transport *domain.Transport) (*bson.ObjectID, *bson.ObjectID) {
	originVesselID := cloneObjectID(schedule.OriginVesselID)
	destinationVesselID := cloneObjectID(schedule.DestinationVesselID)

	if transport != nil {
		if originVesselID == nil {
			originVesselID = cloneObjectID(transport.OriginVesselID)
		}
		if destinationVesselID == nil {
			destinationVesselID = cloneObjectID(transport.DestinationVesselID)
		}
	}

	if schedule.VesselID != nil {
		switch schedule.Direction {
		case domain.TravelDirectionOutbound:
			if originVesselID == nil {
				originVesselID = cloneObjectID(schedule.VesselID)
			}
		case domain.TravelDirectionInbound:
			if destinationVesselID == nil {
				destinationVesselID = cloneObjectID(schedule.VesselID)
			}
		}
	}

	return originVesselID, destinationVesselID
}
