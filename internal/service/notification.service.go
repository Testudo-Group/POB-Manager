package service

import (
	"context"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationService struct {
	repo *repository.NotificationRepository
}

func NewNotificationService(repo *repository.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) CreateNotification(ctx context.Context, n *domain.Notification) error {
	n.ID = bson.NewObjectID()
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	if n.Status == "" {
		n.Status = domain.NotificationStatusPending
	}
	return s.repo.Create(ctx, n)
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userID bson.ObjectID) ([]domain.Notification, error) {
	return s.repo.FindByUserID(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, id bson.ObjectID) error {
	return s.repo.UpdateStatus(ctx, id, domain.NotificationStatusRead)
}
