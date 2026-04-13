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
	CycleAnchorDate time.Time     `json:"cycle_anchor_date" validate:"required"`
}

func (s *RotationService) CreateSchedule(ctx context.Context, input CreateRotationScheduleInput) (*domain.RotationSchedule, error) {
	now := time.Now()
	schedule := &domain.RotationSchedule{
		ID:              bson.NewObjectID(),
		OffshoreRoleID:  input.OffshoreRoleID,
		VesselID:        input.VesselID,
		Name:            input.Name,
		DaysOn:          input.DaysOn,
		DaysOff:         input.DaysOff,
		CycleAnchorDate: input.CycleAnchorDate,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	err := s.scheduleRepo.Create(ctx, schedule)
	if err != nil {
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