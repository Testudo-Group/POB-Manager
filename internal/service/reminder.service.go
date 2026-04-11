package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReminderService struct {
	personnelRepo *repository.PersonnelRepository
	certRepo      *repository.CertificateRepository
	notifSvc      *NotificationService
	userRepo      *repository.UserRepository
}

func NewReminderService(
	pr *repository.PersonnelRepository,
	cr *repository.CertificateRepository,
	ns *NotificationService,
	ur *repository.UserRepository,
) *ReminderService {
	return &ReminderService{
		personnelRepo: pr,
		certRepo:      cr,
		notifSvc:      ns,
		userRepo:      ur,
	}
}

func (s *ReminderService) Start(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour) // Run daily
	go func() {
		for {
			select {
			case <-ticker.C:
				s.CheckExpiringCertificates(ctx)
			case <-ctx.Done():
				ticker.Stop()
				return
			}
		}
	}()
	// Initial run
	s.CheckExpiringCertificates(ctx)
}

func (s *ReminderService) CheckExpiringCertificates(ctx context.Context) {
	log.Println("Starting background certificate expiry check...")
	personnelList, err := s.personnelRepo.FindAll(ctx)
	if err != nil {
		log.Printf("Error fetching personnel for reminders: %v", err)
		return
	}

	for _, p := range personnelList {
		certs, err := s.certRepo.FindByPersonnelID(ctx, p.ID)
		if err != nil {
			continue
		}

		for _, cert := range certs {
			daysUntilExpiry := int(time.Until(cert.ExpiresAt).Hours() / 24)
			
			// Reminders at 6 months (180 days), 4 months (120 days), 1 month (30 days)
			if daysUntilExpiry == 180 || daysUntilExpiry == 120 || daysUntilExpiry == 30 || daysUntilExpiry < 0 {
				s.notify(ctx, p, cert, daysUntilExpiry)
			}
		}
	}
}

func (s *ReminderService) notify(ctx context.Context, p domain.Personnel, cert domain.Certificate, days int) {
	title := "Certificate Expiry Reminder"
	message := fmt.Sprintf("Certificate '%s' for %s %s expires in %d days.", cert.CertificateType, p.FirstName, p.LastName, days)
	if days < 0 {
		title = "Certificate Expired"
		message = fmt.Sprintf("Certificate '%s' for %s %s has EXPIRED.", cert.CertificateType, p.FirstName, p.LastName)
	}

	// 1. Notify the personnel themselves if they have a user account
	if p.UserID != nil {
		s.createNotif(ctx, p.UserID, p.ID, title, message)
	}

	// 2. Notify Safety Admins (Find by role)
	// For simplicity in this iteration, we notify all system_admins as well
	admins, err := s.userRepo.FindAll(ctx) // In real app, filter by "safety_admin" role
	if err == nil {
		for _, admin := range admins {
			if admin.Role == "safety_admin" || admin.Role == "system_admin" {
				s.createNotif(ctx, &admin.ID, p.ID, title, message)
			}
		}
	}
	
	// Stub for email dispatch
	log.Printf("[EMAIL STUB] Sending email: %s - %s", title, message)
}

func (s *ReminderService) createNotif(ctx context.Context, userID *bson.ObjectID, personnelID bson.ObjectID, title, message string) {
	notif := &domain.Notification{
		UserID:            userID,
		PersonnelID:       &personnelID,
		Type:              domain.NotificationTypeComplianceExpiry,
		Channel:           domain.NotificationChannelInApp,
		Title:             title,
		Message:           message,
		Status:            domain.NotificationStatusPending,
		RelatedEntityType: "certificate",
	}
	_ = s.notifSvc.CreateNotification(ctx, notif)
}
