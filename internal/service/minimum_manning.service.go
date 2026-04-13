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
	ErrMinimumManningAlreadyActive = errors.New("minimum manning mode is already active on this vessel")
	ErrMinimumManningNotActive     = errors.New("minimum manning mode is not active on this vessel")
)

type MinimumManningService struct {
	mmRepo          *repository.MinimumManningRepository
	vesselRepo      *repository.VesselRepository
	activityRepo    *repository.ActivityRepository
	personnelRepo   *repository.PersonnelRepository
	roleAssignRepo  *repository.RoleAssignmentRepository
	roleRepo        *repository.OffshoreRoleRepository
	notifSvc        *NotificationService
}

func NewMinimumManningService(
	mmRepo *repository.MinimumManningRepository,
	vesselRepo *repository.VesselRepository,
	activityRepo *repository.ActivityRepository,
	personnelRepo *repository.PersonnelRepository,
	roleAssignRepo *repository.RoleAssignmentRepository,
	roleRepo *repository.OffshoreRoleRepository,
	notifSvc *NotificationService,
) *MinimumManningService {
	return &MinimumManningService{
		mmRepo:         mmRepo,
		vesselRepo:     vesselRepo,
		activityRepo:   activityRepo,
		personnelRepo:  personnelRepo,
		roleAssignRepo: roleAssignRepo,
		roleRepo:       roleRepo,
		notifSvc:       notifSvc,
	}
}

type ActivateMinimumManningInput struct {
	VesselID          bson.ObjectID `json:"vessel_id" validate:"required"`
	ActivatedByUserID bson.ObjectID `json:"activated_by_user_id" validate:"required"`
	Reason            string        `json:"reason"`
}

func (s *MinimumManningService) Activate(ctx context.Context, input ActivateMinimumManningInput) (*domain.MinimumManningEvent, error) {
	activeEvent, err := s.mmRepo.FindActiveByVessel(ctx, input.VesselID)
	if err != nil {
		return nil, err
	}
	if activeEvent != nil {
		return nil, ErrMinimumManningAlreadyActive
	}

	vessel, err := s.vesselRepo.FindByID(ctx, input.VesselID)
	if err != nil {
		return nil, err
	}

	now := time.Now()

	activities, err := s.activityRepo.FindByVessel(ctx, input.VesselID,
		domain.ActivityStatusApproved,
		domain.ActivityStatusSubmitted,
	)
	if err != nil {
		return nil, err
	}

	var affectedActivityIDs []bson.ObjectID
	for _, act := range activities {
		err = s.activityRepo.UpdateStatus(ctx, act.ID, domain.ActivityStatusSuspended, nil, "Minimum manning mode activated")
		if err == nil {
			affectedActivityIDs = append(affectedActivityIDs, act.ID)
		}
	}

	assignments, err := s.roleAssignRepo.FindActiveByVessel(ctx, input.VesselID)
	if err != nil {
		return nil, err
	}

	var affectedPersonnelIDs []bson.ObjectID
	personnelMap := make(map[bson.ObjectID]bool)

	for _, assign := range assignments {
		if !personnelMap[assign.PersonnelID] {
			personnelMap[assign.PersonnelID] = true
			affectedPersonnelIDs = append(affectedPersonnelIDs, assign.PersonnelID)
		}
	}

	vessel.IsMinimumManningActive = true
	vessel.UpdatedAt = now
	err = s.vesselRepo.Update(ctx, vessel)
	if err != nil {
		return nil, err
	}

	event := &domain.MinimumManningEvent{
		ID:                   bson.NewObjectID(),
		VesselID:             input.VesselID,
		ActivatedByUserID:    input.ActivatedByUserID,
		Reason:               input.Reason,
		ReducedPOBCap:        vessel.MinimumSafePOBCapacity,
		Status:               domain.MinimumManningStatusActive,
		ActivatedAt:          now,
		AffectedActivityIDs:  affectedActivityIDs,
		AffectedPersonnelIDs: affectedPersonnelIDs,
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	err = s.mmRepo.Create(ctx, event)
	if err != nil {
		return nil, err
	}

	s.notifyAffectedPersonnel(ctx, event)
	s.notifyActivityOwners(ctx, affectedActivityIDs)

	return event, nil
}

type DeactivateMinimumManningInput struct {
	VesselID            bson.ObjectID `json:"vessel_id" validate:"required"`
	DeactivatedByUserID bson.ObjectID `json:"deactivated_by_user_id" validate:"required"`
	Reason              string        `json:"reason"`
}

func (s *MinimumManningService) Deactivate(ctx context.Context, input DeactivateMinimumManningInput) (*domain.MinimumManningEvent, error) {
	activeEvent, err := s.mmRepo.FindActiveByVessel(ctx, input.VesselID)
	if err != nil {
		return nil, err
	}
	if activeEvent == nil {
		return nil, ErrMinimumManningNotActive
	}

	now := time.Now()

	vessel, err := s.vesselRepo.FindByID(ctx, input.VesselID)
	if err != nil {
		return nil, err
	}
	vessel.IsMinimumManningActive = false
	vessel.UpdatedAt = now
	err = s.vesselRepo.Update(ctx, vessel)
	if err != nil {
		return nil, err
	}

	activeEvent.Status = domain.MinimumManningStatusDeactivated
	activeEvent.DeactivatedByUserID = &input.DeactivatedByUserID
	activeEvent.DeactivatedAt = &now
	activeEvent.UpdatedAt = now

	err = s.mmRepo.Update(ctx, activeEvent)
	if err != nil {
		return nil, err
	}

	s.notifyDeactivation(ctx, activeEvent)

	return activeEvent, nil
}

func (s *MinimumManningService) GetActiveEvent(ctx context.Context, vesselID bson.ObjectID) (*domain.MinimumManningEvent, error) {
	return s.mmRepo.FindActiveByVessel(ctx, vesselID)
}

func (s *MinimumManningService) GetEventHistory(ctx context.Context, vesselID bson.ObjectID, limit int64) ([]domain.MinimumManningEvent, error) {
	return s.mmRepo.FindByVessel(ctx, vesselID, limit)
}

func (s *MinimumManningService) notifyAffectedPersonnel(ctx context.Context, event *domain.MinimumManningEvent) {
	for _, pID := range event.AffectedPersonnelIDs {
		notification := &domain.Notification{
			UserID:    &pID,
			Type:      "minimum_manning_activated",
			Title:     "Minimum Manning Mode Activated",
			Message:   "Minimum manning mode has been activated on your vessel. Non-core activities are suspended.",
			Status:    domain.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = s.notifSvc.CreateNotification(ctx, notification)
	}
}

func (s *MinimumManningService) notifyActivityOwners(ctx context.Context, activityIDs []bson.ObjectID) {
	for _, actID := range activityIDs {
		activity, _ := s.activityRepo.FindByID(ctx, actID)
		if activity != nil {
			userID := activity.CreatedByUserID
			notification := &domain.Notification{
				UserID:    &userID,
				Type:      "activity_suspended",
				Title:     "Activity Suspended - Minimum Manning",
				Message:   "Your activity '" + activity.Name + "' has been suspended due to minimum manning mode activation.",
				Status:    domain.NotificationStatusPending,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}
			_ = s.notifSvc.CreateNotification(ctx, notification)
		}
	}
}

func (s *MinimumManningService) notifyDeactivation(ctx context.Context, event *domain.MinimumManningEvent) {
	for _, pID := range event.AffectedPersonnelIDs {
		notification := &domain.Notification{
			UserID:    &pID,
			Type:      "minimum_manning_deactivated",
			Title:     "Minimum Manning Mode Deactivated",
			Message:   "Minimum manning mode has been deactivated. Normal operations resume.",
			Status:    domain.NotificationStatusPending,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		_ = s.notifSvc.CreateNotification(ctx, notification)
	}
}
