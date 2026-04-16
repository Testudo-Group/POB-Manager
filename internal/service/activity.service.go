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
	ErrActivityConflict       = errors.New("activity scheduling conflict detected")
	ErrInvalidStatusTransition = errors.New("invalid activity status transition")
	ErrPersonnelNotAvailable  = errors.New("personnel not available for activity period")
)

type ActivityService struct {
	activityRepo       *repository.ActivityRepository
	requirementRepo    *repository.ActivityRequirementRepository
	assignmentRepo     *repository.ActivityAssignmentRepository
	roleRepo           *repository.OffshoreRoleRepository
	personnelRepo      *repository.PersonnelRepository
	roleAssignmentRepo *repository.RoleAssignmentRepository
}

func NewActivityService(
	activityRepo *repository.ActivityRepository,
	requirementRepo *repository.ActivityRequirementRepository,
	assignmentRepo *repository.ActivityAssignmentRepository,
	roleRepo *repository.OffshoreRoleRepository,
	personnelRepo *repository.PersonnelRepository,
	roleAssignmentRepo *repository.RoleAssignmentRepository,
) *ActivityService {
	return &ActivityService{
		activityRepo:       activityRepo,
		requirementRepo:    requirementRepo,
		assignmentRepo:     assignmentRepo,
		roleRepo:           roleRepo,
		personnelRepo:      personnelRepo,
		roleAssignmentRepo: roleAssignmentRepo,
	}
}

type ActivityRequirementInput struct {
	OffshoreRoleID bson.ObjectID `json:"offshore_role_id" validate:"required"`
	RequiredCount  int           `json:"required_count" validate:"min=1"`
}

type CreateActivityInput struct {
	VesselID     bson.ObjectID              `json:"vessel_id" validate:"required"`
	Name         string                     `json:"name" validate:"required"`
	Description  string                     `json:"description"`
	StartDate    time.Time                  `json:"start_date" validate:"required"`
	EndDate      time.Time                  `json:"end_date" validate:"required"`
	Priority     domain.ActivityPriority    `json:"priority" validate:"required"`
	CreatedBy    bson.ObjectID              `json:"created_by" validate:"required"`
	Requirements []ActivityRequirementInput `json:"requirements" validate:"required,min=1"`
}

func (s *ActivityService) Create(ctx context.Context, input CreateActivityInput) (*domain.Activity, error) {

	// Validate duration
	durationDays := int(input.EndDate.Sub(input.StartDate).Hours() / 24)
	if durationDays <= 0 {
		return nil, errors.New("end date must be after start date")
	}

	now := time.Now()
	activity := &domain.Activity{
		ID:              bson.NewObjectID(),
		VesselID:        input.VesselID,
		Name:            input.Name,
		Description:     input.Description,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		DurationDays:    durationDays,
		Priority:        input.Priority,
		Status:          domain.ActivityStatusDraft,
		CreatedByUserID: input.CreatedBy,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	       if err := s.activityRepo.Create(ctx, activity); err != nil {
		       return nil, err
	       }

	// Create requirements
	for _, req := range input.Requirements {
		requirement := domain.ActivityRequirement{
			ID:             bson.NewObjectID(),
			ActivityID:     activity.ID,
			OffshoreRoleID: req.OffshoreRoleID,
			RequiredCount:  req.RequiredCount,
		}
		s.requirementRepo.Create(ctx, &requirement)
	}

	return activity, nil
}

func (s *ActivityService) SubmitForApproval(ctx context.Context, activityID, userID bson.ObjectID) error {
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return err
	}

	if activity.Status != domain.ActivityStatusDraft {
		return ErrInvalidStatusTransition
	}

	// Validate requirements exist
	reqs, err := s.requirementRepo.FindByActivity(ctx, activityID)
	if err != nil {
		return err
	}
	if len(reqs) == 0 {
		return errors.New("activity must have at least one role requirement")
	}

	return s.activityRepo.UpdateStatus(ctx, activityID, domain.ActivityStatusSubmitted, nil, "")
}

type ApproveActivityInput struct {
	ActivityID bson.ObjectID `json:"activity_id" validate:"required"`
	ReviewerID bson.ObjectID `json:"reviewer_id" validate:"required"`
	Note       string        `json:"note"`
}

func (s *ActivityService) Approve(ctx context.Context, input ApproveActivityInput) error {
	activity, err := s.activityRepo.FindByID(ctx, input.ActivityID)
	if err != nil {
		return err
	}

	if activity.Status != domain.ActivityStatusSubmitted {
		return ErrInvalidStatusTransition
	}

	return s.activityRepo.UpdateStatus(ctx, input.ActivityID, domain.ActivityStatusApproved, &input.ReviewerID, input.Note)
}

func (s *ActivityService) Reject(ctx context.Context, activityID, reviewerID bson.ObjectID, note string) error {
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return err
	}

	if activity.Status != domain.ActivityStatusSubmitted {
		return ErrInvalidStatusTransition
	}

	return s.activityRepo.UpdateStatus(ctx, activityID, domain.ActivityStatusRejected, &reviewerID, note)
}

func (s *ActivityService) GetPendingApproval(ctx context.Context) ([]domain.Activity, error) {
	return s.activityRepo.FindPendingApproval(ctx)
}

func (s *ActivityService) GetByVessel(ctx context.Context, vesselID bson.ObjectID) ([]domain.Activity, error) {
	return s.activityRepo.FindByVessel(ctx, vesselID)
}

func (s *ActivityService) GetRequirements(ctx context.Context, activityID bson.ObjectID) ([]domain.ActivityRequirement, error) {
	return s.requirementRepo.FindByActivity(ctx, activityID)
}

// Conflict detection
type ActivityConflict struct {
	ActivityID   bson.ObjectID `json:"activity_id"`
	ActivityName string        `json:"activity_name"`
	StartDate    time.Time     `json:"start_date"`
	EndDate      time.Time     `json:"end_date"`
}

func (s *ActivityService) checkConflicts(ctx context.Context, vesselID bson.ObjectID, startDate, endDate time.Time, excludeID *bson.ObjectID) ([]ActivityConflict, error) {
	activities, err := s.activityRepo.FindByDateRange(ctx, vesselID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	var conflicts []ActivityConflict
	for _, a := range activities {
		if excludeID != nil && a.ID == *excludeID {
			continue
		}
		// Only consider approved activities as conflicts
		if a.Status == domain.ActivityStatusApproved {
			conflicts = append(conflicts, ActivityConflict{
				ActivityID:   a.ID,
				ActivityName: a.Name,
				StartDate:    a.StartDate,
				EndDate:      a.EndDate,
			})
		}
	}
	return conflicts, nil
}

func (s *ActivityService) CheckConflicts(ctx context.Context, activityID bson.ObjectID) ([]ActivityConflict, error) {
	activity, err := s.activityRepo.FindByID(ctx, activityID)
	if err != nil {
		return nil, err
	}
	return s.checkConflicts(ctx, activity.VesselID, activity.StartDate, activity.EndDate, &activityID)
}

// Gantt Chart Data
type GanttActivity struct {
	ID           bson.ObjectID                 `json:"id"`
	Name         string                        `json:"name"`
	StartDate    time.Time                     `json:"start_date"`
	EndDate      time.Time                     `json:"end_date"`
	DurationDays int                           `json:"duration_days"`
	Priority     domain.ActivityPriority       `json:"priority"`
	Status       domain.ActivityStatus         `json:"status"`
	Requirements []domain.ActivityRequirement  `json:"requirements"`
	Assignments  []domain.ActivityAssignment   `json:"assignments"`
}

func (s *ActivityService) GetGanttData(ctx context.Context, vesselID bson.ObjectID) ([]GanttActivity, error) {
	activities, err := s.activityRepo.FindByVessel(ctx, vesselID,
		domain.ActivityStatusApproved,
		domain.ActivityStatusSubmitted,
	)
	if err != nil {
		return nil, err
	}

	var ganttData []GanttActivity
	for _, a := range activities {
		reqs, _ := s.requirementRepo.FindByActivity(ctx, a.ID)
		assignments, _ := s.assignmentRepo.FindByActivity(ctx, a.ID)

		ganttData = append(ganttData, GanttActivity{
			ID:           a.ID,
			Name:         a.Name,
			StartDate:    a.StartDate,
			EndDate:      a.EndDate,
			DurationDays: a.DurationDays,
			Priority:     a.Priority,
			Status:       a.Status,
			Requirements: reqs,
			Assignments:  assignments,
		})
	}
	return ganttData, nil
}

// Personnel Assignment
type AssignPersonnelInput struct {
	ActivityID   bson.ObjectID   `json:"activity_id" validate:"required"`
	PersonnelIDs []bson.ObjectID `json:"personnel_ids" validate:"required"`
}

func (s *ActivityService) AssignPersonnel(ctx context.Context, input AssignPersonnelInput) error {
	activity, err := s.activityRepo.FindByID(ctx, input.ActivityID)
	if err != nil {
		return err
	}

	if activity.Status != domain.ActivityStatusApproved {
		return errors.New("can only assign personnel to approved activities")
	}

	now := time.Now()
	var assignments []domain.ActivityAssignment
	for _, pID := range input.PersonnelIDs {
		// Check personnel availability
		existing, _ := s.assignmentRepo.FindByPersonnel(ctx, pID)
		for _, a := range existing {
			existingActivity, _ := s.activityRepo.FindByID(ctx, a.ActivityID)
			if existingActivity != nil {
				if existingActivity.StartDate.Before(activity.EndDate) && existingActivity.EndDate.After(activity.StartDate) {
					return ErrPersonnelNotAvailable
				}
			}
		}

		assignments = append(assignments, domain.ActivityAssignment{
			ID:          bson.NewObjectID(),
			ActivityID:  input.ActivityID,
			PersonnelID: pID,
			Status:      "assigned",
			CreatedAt:   now,
			UpdatedAt:   now,
		})
	}

	return s.assignmentRepo.CreateMany(ctx, assignments)
}

func (s *ActivityService) GetAssignments(ctx context.Context, activityID bson.ObjectID) ([]domain.ActivityAssignment, error) {
	return s.assignmentRepo.FindByActivity(ctx, activityID)
}

func (s *ActivityService) Delete(ctx context.Context, id bson.ObjectID) error {
	activity, err := s.activityRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if activity.Status == domain.ActivityStatusApproved {
		return errors.New("cannot delete approved activity")
	}

	// Delete requirements and assignments
	s.requirementRepo.DeleteByActivity(ctx, id)
	s.assignmentRepo.DeleteByActivity(ctx, id)

	return s.activityRepo.Delete(ctx, id)
}
func (s *ActivityService) GetByID(ctx context.Context, id bson.ObjectID) (*domain.Activity, error) {
	return s.activityRepo.FindByID(ctx, id)
}
