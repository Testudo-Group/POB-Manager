package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	ErrRoleAlreadyAssigned = errors.New("personnel already has active assignment for this role")
	ErrMinimumRoleCount    = errors.New("cannot remove assignment: minimum role count requirement not met")
)

type RotationService struct {
	scheduleRepo     *repository.RotationScheduleRepository
	assignmentRepo   *repository.RoleAssignmentRepository
	backToBackRepo   *repository.BackToBackPairRepository
	roleRepo         *repository.OffshoreRoleRepository
	personnelRepo    *repository.PersonnelRepository
	roomAssignRepo   *repository.RoomAssignmentRepository
}

func NewRotationService(
	scheduleRepo *repository.RotationScheduleRepository,
	assignmentRepo *repository.RoleAssignmentRepository,
	backToBackRepo *repository.BackToBackPairRepository,
	roleRepo *repository.OffshoreRoleRepository,
	personnelRepo *repository.PersonnelRepository,
	roomAssignRepo *repository.RoomAssignmentRepository,
) *RotationService {
	return &RotationService{
		scheduleRepo:   scheduleRepo,
		assignmentRepo: assignmentRepo,
		backToBackRepo: backToBackRepo,
		roleRepo:       roleRepo,
		personnelRepo:  personnelRepo,
		roomAssignRepo: roomAssignRepo,
	}
}

// Rotation Schedule
type CreateRotationScheduleInput struct {
	OffshoreRoleID  bson.ObjectID `json:"offshore_role_id" validate:"required"`
	VesselID        bson.ObjectID `json:"vessel_id" validate:"required"`
	Name            string        `json:"name" validate:"required"`
	DaysOn          int           `json:"days_on" validate:"required,min=1"`
	DaysOff         int           `json:"days_off" validate:"required,min=1"`
	CycleAnchorDate string        `json:"cycle_anchor_date" validate:"required"`
}

func parseFlexibleDate(s string) (time.Time, error) {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("cannot parse date %q", s)
}

func (s *RotationService) CreateSchedule(ctx context.Context, input CreateRotationScheduleInput) (*domain.RotationSchedule, error) {
	anchorDate, err := parseFlexibleDate(input.CycleAnchorDate)
	if err != nil {
		return nil, fmt.Errorf("invalid cycle_anchor_date: %w", err)
	}

	now := time.Now()
	schedule := &domain.RotationSchedule{
		ID:              bson.NewObjectID(),
		OffshoreRoleID:  input.OffshoreRoleID,
		VesselID:        input.VesselID,
		Name:            input.Name,
		DaysOn:          input.DaysOn,
		DaysOff:         input.DaysOff,
		CycleAnchorDate: anchorDate,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err = s.scheduleRepo.Create(ctx, schedule); err != nil {
		return nil, err
	}
	return schedule, nil
}

func (s *RotationService) GetSchedulesByRole(ctx context.Context, roleID, vesselID bson.ObjectID) ([]domain.RotationSchedule, error) {
	return s.scheduleRepo.FindByRoleAndVessel(ctx, roleID, vesselID)
}

// Role Assignment
type AssignRoleInput struct {
	OffshoreRoleID     bson.ObjectID  `json:"offshore_role_id" validate:"required"`
	PersonnelID        bson.ObjectID  `json:"personnel_id" validate:"required"`
	VesselID           bson.ObjectID  `json:"vessel_id" validate:"required"`
	RotationScheduleID *bson.ObjectID `json:"rotation_schedule_id"`
	RoomID             *bson.ObjectID `json:"room_id"`
	AssignedByUserID   bson.ObjectID  `json:"assigned_by_user_id" validate:"required"`
	EffectiveFrom      time.Time      `json:"effective_from" validate:"required"`
}

func (s *RotationService) AssignRole(ctx context.Context, input AssignRoleInput) (*domain.RoleAssignment, error) {
	// Check if personnel already has active assignment for this role on this vessel
	existing, _ := s.assignmentRepo.FindActiveByPersonnel(ctx, input.PersonnelID)
	for _, assign := range existing {
		if assign.OffshoreRoleID == input.OffshoreRoleID && assign.VesselID == input.VesselID {
			return nil, ErrRoleAlreadyAssigned
		}
	}

	now := time.Now()
	assignment := &domain.RoleAssignment{
		ID:                 bson.NewObjectID(),
		OffshoreRoleID:     input.OffshoreRoleID,
		PersonnelID:        input.PersonnelID,
		VesselID:           input.VesselID,
		RotationScheduleID: input.RotationScheduleID,
		RoomID:             input.RoomID,
		AssignedByUserID:   &input.AssignedByUserID,
		EffectiveFrom:      input.EffectiveFrom,
		Status:             domain.RoleAssignmentStatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err := s.assignmentRepo.Create(ctx, assignment)
	if err != nil {
		return nil, err
	}

	// Update personnel status to onboard
	s.personnelRepo.Update(ctx, &domain.Personnel{
		ID:            input.PersonnelID,
		CurrentStatus: domain.PersonnelStatusOnboard,
		UpdatedAt:     now,
	})

	return assignment, nil
}

func (s *RotationService) EndAssignment(ctx context.Context, assignmentID bson.ObjectID, effectiveTo time.Time) error {
	assignment, err := s.assignmentRepo.FindByID(ctx, assignmentID)
	if err != nil {
		return err
	}

	// Check minimum role count before ending
	role, err := s.roleRepo.FindByID(ctx, assignment.OffshoreRoleID)
	if err != nil {
		return err
	}

	if role.MinimumRequiredCount > 0 {
		count, err := s.assignmentRepo.CountActiveByRole(ctx, assignment.OffshoreRoleID, assignment.VesselID)
		if err != nil {
			return err
		}
		if int(count) <= role.MinimumRequiredCount {
			return ErrMinimumRoleCount
		}
	}

	assignment.Status = domain.RoleAssignmentStatusCompleted
	assignment.EffectiveTo = &effectiveTo
	assignment.UpdatedAt = time.Now()

	err = s.assignmentRepo.Update(ctx, assignment)
	if err != nil {
		return err
	}

	// Check if personnel has other active assignments
	activeAssignments, _ := s.assignmentRepo.FindActiveByPersonnel(ctx, assignment.PersonnelID)
	if len(activeAssignments) == 0 {
		s.personnelRepo.Update(ctx, &domain.Personnel{
			ID:            assignment.PersonnelID,
			CurrentStatus: domain.PersonnelStatusAvailable,
			UpdatedAt:     time.Now(),
		})
	}

	return nil
}

func (s *RotationService) GetVesselManning(ctx context.Context, vesselID bson.ObjectID) ([]domain.RoleAssignment, error) {
	return s.assignmentRepo.FindActiveByVessel(ctx, vesselID)
}

// Back-to-Back Pair
type CreateBackToBackPairInput struct {
	OffshoreRoleID     bson.ObjectID  `json:"offshore_role_id" validate:"required"`
	VesselID           bson.ObjectID  `json:"vessel_id" validate:"required"`
	PrimaryPersonnelID bson.ObjectID  `json:"primary_personnel_id" validate:"required"`
	ReliefPersonnelID  bson.ObjectID  `json:"relief_personnel_id" validate:"required"`
	RoomID             *bson.ObjectID `json:"room_id"`
	Notes              string         `json:"notes"`
	EffectiveFrom      time.Time      `json:"effective_from" validate:"required"`
}

func (s *RotationService) CreateBackToBackPair(ctx context.Context, input CreateBackToBackPairInput) (*domain.BackToBackPair, error) {
	now := time.Now()
	pair := &domain.BackToBackPair{
		ID:                 bson.NewObjectID(),
		OffshoreRoleID:     input.OffshoreRoleID,
		VesselID:           input.VesselID,
		PrimaryPersonnelID: input.PrimaryPersonnelID,
		ReliefPersonnelID:  input.ReliefPersonnelID,
		RoomID:             input.RoomID,
		Notes:              input.Notes,
		Status:             domain.BackToBackPairStatusActive,
		EffectiveFrom:      input.EffectiveFrom,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	err := s.backToBackRepo.Create(ctx, pair)
	if err != nil {
		return nil, err
	}
	return pair, nil
}

func (s *RotationService) GetBackToBackPairsByRole(ctx context.Context, roleID, vesselID bson.ObjectID) ([]domain.BackToBackPair, error) {
	return s.backToBackRepo.FindActiveByRole(ctx, roleID, vesselID)
}

var ErrNoPairForHandover = errors.New("no active back-to-back pair found for this assignment")

type ShiftHandoverResult struct {
	EndedAssignment *domain.RoleAssignment `json:"ended_assignment"`
	NewAssignment   *domain.RoleAssignment `json:"new_assignment"`
	ReliefPersonnel bson.ObjectID          `json:"relief_personnel_id"`
}

// TriggerShiftHandover ends the current personnel's assignment and activates the relief personnel.
func (s *RotationService) TriggerShiftHandover(ctx context.Context, assignmentID bson.ObjectID, handoverAt time.Time) (*ShiftHandoverResult, error) {
	assignment, err := s.assignmentRepo.FindByID(ctx, assignmentID)
	if err != nil {
		return nil, err
	}

	// Find back-to-back pair where outgoing personnel is primary
	pairs, err := s.backToBackRepo.FindByPersonnel(ctx, assignment.PersonnelID)
	if err != nil {
		return nil, err
	}

	var activePair *domain.BackToBackPair
	for i := range pairs {
		p := &pairs[i]
		if p.OffshoreRoleID == assignment.OffshoreRoleID && p.VesselID == assignment.VesselID {
			activePair = p
			break
		}
	}

	if activePair == nil {
		return nil, ErrNoPairForHandover
	}

	reliefID := activePair.ReliefPersonnelID
	if activePair.PrimaryPersonnelID != assignment.PersonnelID {
		reliefID = activePair.PrimaryPersonnelID
	}

	// End current assignment
	assignment.Status = domain.RoleAssignmentStatusCompleted
	assignment.EffectiveTo = &handoverAt
	assignment.UpdatedAt = time.Now()
	if err := s.assignmentRepo.Update(ctx, assignment); err != nil {
		return nil, err
	}

	// Create new assignment for relief personnel
	now := time.Now()
	newAssignment := &domain.RoleAssignment{
		ID:                 bson.NewObjectID(),
		OffshoreRoleID:     assignment.OffshoreRoleID,
		PersonnelID:        reliefID,
		VesselID:           assignment.VesselID,
		RotationScheduleID: assignment.RotationScheduleID,
		RoomID:             assignment.RoomID,
		AssignedByUserID:   assignment.AssignedByUserID,
		EffectiveFrom:      handoverAt,
		Status:             domain.RoleAssignmentStatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}
	if err := s.assignmentRepo.Create(ctx, newAssignment); err != nil {
		return nil, err
	}

	// Swap primary and relief in the pair
	activePair.PrimaryPersonnelID = reliefID
	activePair.ReliefPersonnelID = assignment.PersonnelID
	activePair.UpdatedAt = now
	_ = s.backToBackRepo.Update(ctx, activePair)

	// Update personnel statuses
	s.personnelRepo.Update(ctx, &domain.Personnel{
		ID:            reliefID,
		CurrentStatus: domain.PersonnelStatusOnboard,
		UpdatedAt:     now,
	})

	activeAssignments, _ := s.assignmentRepo.FindActiveByPersonnel(ctx, assignment.PersonnelID)
	if len(activeAssignments) == 0 {
		s.personnelRepo.Update(ctx, &domain.Personnel{
			ID:            assignment.PersonnelID,
			CurrentStatus: domain.PersonnelStatusAvailable,
			UpdatedAt:     now,
		})
	}

	return &ShiftHandoverResult{
		EndedAssignment: assignment,
		NewAssignment:   newAssignment,
		ReliefPersonnel: reliefID,
	}, nil
}

// Calculate next rotation dates
func (s *RotationService) CalculateNextRotation(ctx context.Context, scheduleID bson.ObjectID, fromDate time.Time) (time.Time, time.Time, error) {
	schedule, err := s.scheduleRepo.FindByID(ctx, scheduleID)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}

	daysSinceAnchor := int(fromDate.Sub(schedule.CycleAnchorDate).Hours() / 24)
	cycleLength := schedule.DaysOn + schedule.DaysOff
	_ = daysSinceAnchor / cycleLength
	daysIntoCurrentCycle := daysSinceAnchor % cycleLength

	var nextOnDate, nextOffDate time.Time

	if daysIntoCurrentCycle < schedule.DaysOn {
		// Currently in ON period
		nextOffDate = fromDate.AddDate(0, 0, schedule.DaysOn-daysIntoCurrentCycle)
		nextOnDate = nextOffDate.AddDate(0, 0, schedule.DaysOff)
	} else {
		// Currently in OFF period
		daysIntoOff := daysIntoCurrentCycle - schedule.DaysOn
		nextOnDate = fromDate.AddDate(0, 0, schedule.DaysOff-daysIntoOff)
		nextOffDate = nextOnDate.AddDate(0, 0, schedule.DaysOn)
	}

	return nextOnDate, nextOffDate, nil
}

// ActiveAssignmentSummary is the enriched view returned to the frontend.
type ActiveAssignmentSummary struct {
	AssignmentID   string `json:"assignment_id"`
	PersonnelID    string `json:"personnel_id"`
	PersonnelName  string `json:"personnel_name"`
	RoleID         string `json:"role_id"`
	RoleName       string `json:"role_name"`
	VesselID       string `json:"vessel_id"`
	EffectiveFrom  string `json:"effective_from"`
}

func (s *RotationService) GetActiveAssignmentsForVessel(ctx context.Context, vesselID bson.ObjectID) ([]ActiveAssignmentSummary, error) {
	assignments, err := s.assignmentRepo.FindActiveByVessel(ctx, vesselID)
	if err != nil {
		return nil, err
	}

	result := make([]ActiveAssignmentSummary, 0, len(assignments))
	for _, a := range assignments {
		summary := ActiveAssignmentSummary{
			AssignmentID:  a.ID.Hex(),
			PersonnelID:   a.PersonnelID.Hex(),
			RoleID:        a.OffshoreRoleID.Hex(),
			VesselID:      a.VesselID.Hex(),
			EffectiveFrom: a.EffectiveFrom.Format("2006-01-02"),
		}

		if p, err := s.personnelRepo.FindByID(ctx, a.PersonnelID); err == nil {
			summary.PersonnelName = p.FirstName + " " + p.LastName
		}
		if r, err := s.roleRepo.FindByID(ctx, a.OffshoreRoleID); err == nil {
			summary.RoleName = r.Name
		}

		result = append(result, summary)
	}
	return result, nil
}

// EnrichedBackToBackPair resolves personnel and role names in a B2B pair.
type EnrichedBackToBackPair struct {
	ID               string `json:"id"`
	PrimaryID        string `json:"primary_personnel_id"`
	PrimaryName      string `json:"primary_personnel_name"`
	ReliefID         string `json:"relief_personnel_id"`
	ReliefName       string `json:"relief_personnel_name"`
	RoleID           string `json:"role_id"`
	RoleName         string `json:"role_name"`
	VesselID         string `json:"vessel_id"`
}

func (s *RotationService) GetEnrichedBackToBackPairs(ctx context.Context, roleID, vesselID bson.ObjectID) ([]EnrichedBackToBackPair, error) {
	pairs, err := s.backToBackRepo.FindActiveByRole(ctx, roleID, vesselID)
	if err != nil {
		return nil, err
	}

	result := make([]EnrichedBackToBackPair, 0, len(pairs))
	for _, p := range pairs {
		ep := EnrichedBackToBackPair{
			ID:       p.ID.Hex(),
			PrimaryID: p.PrimaryPersonnelID.Hex(),
			ReliefID:  p.ReliefPersonnelID.Hex(),
			RoleID:    p.OffshoreRoleID.Hex(),
			VesselID:  p.VesselID.Hex(),
		}
		if person, err := s.personnelRepo.FindByID(ctx, p.PrimaryPersonnelID); err == nil {
			ep.PrimaryName = person.FirstName + " " + person.LastName
		}
		if person, err := s.personnelRepo.FindByID(ctx, p.ReliefPersonnelID); err == nil {
			ep.ReliefName = person.FirstName + " " + person.LastName
		}
		if role, err := s.roleRepo.FindByID(ctx, p.OffshoreRoleID); err == nil {
			ep.RoleName = role.Name
		}
		result = append(result, ep)
	}
	return result, nil
}