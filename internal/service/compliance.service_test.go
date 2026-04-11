package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/service"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// --- Mocks ---

type mockPersonnelReader struct {
	personnel *domain.Personnel
	err       error
}

func (m *mockPersonnelReader) FindByID(ctx context.Context, id bson.ObjectID) (*domain.Personnel, error) {
	return m.personnel, m.err
}

type mockRoleReader struct {
	roles map[bson.ObjectID]*domain.OffshoreRole
}

func (m *mockRoleReader) FindByID(ctx context.Context, id bson.ObjectID) (*domain.OffshoreRole, error) {
	if role, ok := m.roles[id]; ok {
		return role, nil
	}
	return nil, nil // Or an error, but nil err mimics "not found ignored" in service
}

type mockCertTypeReader struct {
	types map[bson.ObjectID]*domain.CertificateType
}

func (m *mockCertTypeReader) FindByID(ctx context.Context, id bson.ObjectID) (*domain.CertificateType, error) {
	if ct, ok := m.types[id]; ok {
		return ct, nil
	}
	return nil, nil
}

type mockCertReader struct {
	certs []domain.Certificate
	err   error
}

func (m *mockCertReader) FindByPersonnelID(ctx context.Context, personnelID bson.ObjectID) ([]domain.Certificate, error) {
	return m.certs, m.err
}

// --- Tests ---

func TestComplianceService_CheckCompliance(t *testing.T) {
	personnelID := bson.NewObjectID()
	roleID := bson.NewObjectID()
	certTypeID := bson.NewObjectID()
	now := time.Now()

	tests := []struct {
		name                 string
		setupMocks           func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader)
		expectedStatus       string
		expectedMissing      int
		expectedExpired      int
		expectedExpiring     int
	}{
		{
			name: "No Roles Assigned - Should be Non-Compliant",
			setupMocks: func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader) {
				return &mockPersonnelReader{
						personnel: &domain.Personnel{
							ID:              personnelID,
							OffshoreRoleIDs: []bson.ObjectID{}, // Empty
						},
					},
					&mockRoleReader{},
					&mockCertReader{},
					&mockCertTypeReader{}
			},
			expectedStatus:  "Non-Compliant (No Role Assigned)",
			expectedMissing: 1, // "NO_ROLE_ASSIGNED"
		},
		{
			name: "Fully Compliant - Has Required Certificate",
			setupMocks: func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader) {
				return &mockPersonnelReader{
						personnel: &domain.Personnel{
							ID:              personnelID,
							OffshoreRoleIDs: []bson.ObjectID{roleID},
						},
					},
					&mockRoleReader{
						roles: map[bson.ObjectID]*domain.OffshoreRole{
							roleID: {
								ID:                       roleID,
								RequiredCertificateTypes: []bson.ObjectID{certTypeID},
							},
						},
					},
					&mockCertReader{
						certs: []domain.Certificate{
							{
								CertificateType: "BOSIET",
								ExpiresAt:       now.Add(60 * 24 * time.Hour), // 60 days
								Status:          domain.CertificateStatusValid,
							},
						},
					},
					&mockCertTypeReader{
						types: map[bson.ObjectID]*domain.CertificateType{
							certTypeID: {Code: "BOSIET"},
						},
					}
			},
			expectedStatus: "Compliant",
		},
		{
			name: "Missing Certificate",
			setupMocks: func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader) {
				return &mockPersonnelReader{
						personnel: &domain.Personnel{
							ID:              personnelID,
							OffshoreRoleIDs: []bson.ObjectID{roleID},
						},
					},
					&mockRoleReader{
						roles: map[bson.ObjectID]*domain.OffshoreRole{
							roleID: {
								ID:                       roleID,
								RequiredCertificateTypes: []bson.ObjectID{certTypeID},
							},
						},
					},
					&mockCertReader{
						certs: []domain.Certificate{}, // Empty, none uploaded
					},
					&mockCertTypeReader{
						types: map[bson.ObjectID]*domain.CertificateType{
							certTypeID: {Code: "BOSIET"},
						},
					}
			},
			expectedStatus:  "Non-Compliant",
			expectedMissing: 1,
		},
		{
			name: "Expired Certificate",
			setupMocks: func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader) {
				return &mockPersonnelReader{
						personnel: &domain.Personnel{
							ID:              personnelID,
							OffshoreRoleIDs: []bson.ObjectID{roleID},
						},
					},
					&mockRoleReader{
						roles: map[bson.ObjectID]*domain.OffshoreRole{
							roleID: {
								ID:                       roleID,
								RequiredCertificateTypes: []bson.ObjectID{certTypeID},
							},
						},
					},
					&mockCertReader{
						certs: []domain.Certificate{
							{
								CertificateType: "BOSIET",
								ExpiresAt:       now.Add(-2 * 24 * time.Hour), // Expired 2 days ago
								Status:          domain.CertificateStatusExpired,
							},
						},
					},
					&mockCertTypeReader{
						types: map[bson.ObjectID]*domain.CertificateType{
							certTypeID: {Code: "BOSIET"},
						},
					}
			},
			expectedStatus:   "Non-Compliant",
			expectedMissing:  1, // Counted as missing because it's expired and rejected from active map
			expectedExpired:  1,
		},
		{
			name: "Expiring Soon Certificate - Still Compliant",
			setupMocks: func() (*mockPersonnelReader, *mockRoleReader, *mockCertReader, *mockCertTypeReader) {
				return &mockPersonnelReader{
						personnel: &domain.Personnel{
							ID:              personnelID,
							OffshoreRoleIDs: []bson.ObjectID{roleID},
						},
					},
					&mockRoleReader{
						roles: map[bson.ObjectID]*domain.OffshoreRole{
							roleID: {
								ID:                       roleID,
								RequiredCertificateTypes: []bson.ObjectID{certTypeID},
							},
						},
					},
					&mockCertReader{
						certs: []domain.Certificate{
							{
								CertificateType: "BOSIET",
								ExpiresAt:       now.Add(10 * 24 * time.Hour), // Expiring in 10 days
								Status:          domain.CertificateStatusValid,
							},
						},
					},
					&mockCertTypeReader{
						types: map[bson.ObjectID]*domain.CertificateType{
							certTypeID: {Code: "BOSIET"},
						},
					}
			},
			expectedStatus:   "Compliant",
			expectedExpiring: 1, // Will be flagged for the basic email alert!
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, r, c, ct := tt.setupMocks()
			svc := service.NewComplianceService(p, r, c, ct)

			result, err := svc.CheckCompliance(context.Background(), personnelID)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Status != tt.expectedStatus {
				t.Errorf("expected status %q, got %q", tt.expectedStatus, result.Status)
			}
			if len(result.MissingCertificates) != tt.expectedMissing {
				t.Errorf("expected %d missing certs, got %d", tt.expectedMissing, len(result.MissingCertificates))
			}
			if len(result.ExpiredCertificates) != tt.expectedExpired {
				t.Errorf("expected %d expired certs, got %d", tt.expectedExpired, len(result.ExpiredCertificates))
			}
			if len(result.ExpiringCertificates) != tt.expectedExpiring {
				t.Errorf("expected %d expiring certs, got %d", tt.expectedExpiring, len(result.ExpiringCertificates))
			}
		})
	}
}
