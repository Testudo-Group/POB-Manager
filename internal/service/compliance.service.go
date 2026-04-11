package service

import (
	"context"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PersonnelReader interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.Personnel, error)
}

type OffshoreRoleReader interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.OffshoreRole, error)
}

type CertificateReader interface {
	FindByPersonnelID(ctx context.Context, personnelID bson.ObjectID) ([]domain.Certificate, error)
}

type CertificateTypeReader interface {
	FindByID(ctx context.Context, id bson.ObjectID) (*domain.CertificateType, error)
}

type ComplianceService struct {
	personnelRepo PersonnelReader
	roleRepo      OffshoreRoleReader
	certRepo      CertificateReader
	typeRepo      CertificateTypeReader
}

func NewComplianceService(
	p PersonnelReader,
	r OffshoreRoleReader,
	c CertificateReader,
	t CertificateTypeReader,
) *ComplianceService {
	return &ComplianceService{
		personnelRepo: p,
		roleRepo:      r,
		certRepo:      c,
		typeRepo:      t,
	}
}

type ComplianceResult struct {
	PersonnelID          bson.ObjectID `json:"personnel_id"`
	Status               string        `json:"status"` // Compliant, Non-Compliant
	MissingCertificates  []string      `json:"missing_certificates"`
	ExpiredCertificates  []string      `json:"expired_certificates"`
	ExpiringCertificates []string      `json:"expiring_certificates"`
}

func (s *ComplianceService) CheckCompliance(ctx context.Context, personnelID bson.ObjectID) (*ComplianceResult, error) {
	// 1. Get Personnel
	personnel, err := s.personnelRepo.FindByID(ctx, personnelID)
	if err != nil {
		return nil, err
	}

	if len(personnel.OffshoreRoleIDs) == 0 {
		return &ComplianceResult{
			PersonnelID:          personnelID,
			Status:               "Non-Compliant (No Role Assigned)",
			MissingCertificates:  []string{"NO_ROLE_ASSIGNED"},
			ExpiredCertificates:  []string{},
			ExpiringCertificates: []string{},
		}, nil
	}

	// 2. Aggregate Required Certificate Types
	requiredTypeMap := make(map[bson.ObjectID]bool)
	for _, roleID := range personnel.OffshoreRoleIDs {
		role, err := s.roleRepo.FindByID(ctx, roleID)
		if err == nil {
			for _, ctID := range role.RequiredCertificateTypes {
				requiredTypeMap[ctID] = true
			}
		}
	}

	// Translate ObjectIDs to Codes (for checking string matching on Certificate.CertificateType)
	requiredCodesMap := make(map[string]bool)
	for ctID := range requiredTypeMap {
		ct, err := s.typeRepo.FindByID(ctx, ctID)
		if err == nil {
			requiredCodesMap[ct.Code] = true
		}
	}

	// 3. Get User's Certificates
	userCerts, err := s.certRepo.FindByPersonnelID(ctx, personnelID)
	if err != nil {
		return nil, err
	}

	activeCertsMap := make(map[string]bool)
	var expired []string
	var expiring []string
	now := time.Now()

	for _, cert := range userCerts {
		if now.After(cert.ExpiresAt) {
			expired = append(expired, cert.CertificateType)
			continue
		}
		activeCertsMap[cert.CertificateType] = true
		
		// Check expiring soon (< 30 days)
		if cert.ExpiresAt.Sub(now) < 30*24*time.Hour {
			expiring = append(expiring, cert.CertificateType)
		}
	}

	// 4. Determine Missing
	var missing []string
	for code := range requiredCodesMap {
		if !activeCertsMap[code] {
			missing = append(missing, code)
		}
	}

	status := "Compliant"
	if len(missing) > 0 || len(expired) > 0 {
		status = "Non-Compliant"
	}

	return &ComplianceResult{
		PersonnelID:          personnelID,
		Status:               status,
		MissingCertificates:  missing,
		ExpiredCertificates:  expired,
		ExpiringCertificates: expiring,
	}, nil
}
