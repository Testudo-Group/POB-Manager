package service

import (
	"context"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DashboardService struct {
	vesselRepo     *repository.VesselRepository
	activityRepo   *repository.ActivityRepository
	certRepo       *repository.CertificateRepository
	transportRepo  *repository.TransportRepository
	travelSchedRepo *repository.TravelScheduleRepository
	travelAssignRepo *repository.TravelAssignmentRepository
	personnelRepo  *repository.PersonnelRepository
	roomAssignRepo *repository.RoomAssignmentRepository
	vesselSvc      *VesselService
}

func NewDashboardService(
	vesselRepo *repository.VesselRepository,
	activityRepo *repository.ActivityRepository,
	certRepo *repository.CertificateRepository,
	transportRepo *repository.TransportRepository,
	travelSchedRepo *repository.TravelScheduleRepository,
	travelAssignRepo *repository.TravelAssignmentRepository,
	personnelRepo *repository.PersonnelRepository,
	roomAssignRepo *repository.RoomAssignmentRepository,
	vesselSvc *VesselService,
) *DashboardService {
	return &DashboardService{
		vesselRepo:      vesselRepo,
		activityRepo:    activityRepo,
		certRepo:        certRepo,
		transportRepo:   transportRepo,
		travelSchedRepo: travelSchedRepo,
		travelAssignRepo: travelAssignRepo,
		personnelRepo:   personnelRepo,
		roomAssignRepo:  roomAssignRepo,
		vesselSvc:       vesselSvc,
	}
}

type DashboardData struct {
	RealTimePOB          POBWidget             `json:"real_time_pob"`
	UpcomingActivities   []UpcomingActivity    `json:"upcoming_activities"`
	ExpiringCertificates []ExpiringCertificate `json:"expiring_certificates"`
	UpcomingTravel       []UpcomingTravel      `json:"upcoming_travel"`
	ActivityPriorityDist map[string]int        `json:"activity_priority_distribution"`
}

type POBWidget struct {
	VesselID         string  `json:"vessel_id"`
	VesselName       string  `json:"vessel_name"`
	CurrentPOB       int     `json:"current_pob"`
	Capacity         int     `json:"capacity"`
	Utilization      float64 `json:"utilization_percent"`
	MinManningActive bool    `json:"min_manning_active"`
}

type UpcomingActivity struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	StartDate  time.Time `json:"start_date"`
	EndDate    time.Time `json:"end_date"`
	Priority   string    `json:"priority"`
	Status     string    `json:"status"`
	VesselName string    `json:"vessel_name"`
}

type ExpiringCertificate struct {
	PersonnelID     string    `json:"personnel_id"`
	PersonnelName   string    `json:"personnel_name"`
	CertificateType string    `json:"certificate_type"`
	ExpiryDate      time.Time `json:"expiry_date"`
	DaysRemaining   int       `json:"days_remaining"`
}

type UpcomingTravel struct {
	ScheduleID    string    `json:"schedule_id"`
	TransportName string    `json:"transport_name"`
	Direction     string    `json:"direction"`
	DepartureAt   time.Time `json:"departure_at"`
	VesselName    string    `json:"vessel_name,omitempty"`
	SeatsBooked   int       `json:"seats_booked"`
	Capacity      int       `json:"capacity"`
}

func (s *DashboardService) GetRoleFilteredDashboard(ctx context.Context, userID bson.ObjectID, userRole domain.UserRole, vesselID *bson.ObjectID) (*DashboardData, error) {
	data := &DashboardData{
		UpcomingActivities:   []UpcomingActivity{},
		ExpiringCertificates: []ExpiringCertificate{},
		UpcomingTravel:       []UpcomingTravel{},
		ActivityPriorityDist: make(map[string]int),
	}

	// Determine target vessel (default to first active vessel if not specified)
	var targetVesselID bson.ObjectID
	if vesselID != nil && !vesselID.IsZero() {
		targetVesselID = *vesselID
	} else {
		vessels, err := s.vesselRepo.FindAll(ctx)
		if err == nil && len(vessels) > 0 {
			for _, v := range vessels {
				if v.Status == domain.VesselStatusActive {
					targetVesselID = v.ID
					break
				}
			}
		}
	}

	// 1. Real-time POB Widget
	if !targetVesselID.IsZero() {
		vessel, err := s.vesselRepo.FindByID(ctx, targetVesselID)
		if err == nil {
			pob, _ := s.vesselSvc.GetRealTimePOB(ctx, targetVesselID)
			utilization := 0.0
			if vessel.POBCapacity > 0 {
				utilization = float64(pob) / float64(vessel.POBCapacity) * 100
			}
			data.RealTimePOB = POBWidget{
				VesselID:         targetVesselID.Hex(),
				VesselName:       vessel.Name,
				CurrentPOB:       pob,
				Capacity:         vessel.POBCapacity,
				Utilization:      utilization,
				MinManningActive: vessel.IsMinimumManningActive,
			}
		}
	}

	// 2. Upcoming Activities (next 7 days) - for roles with view access
	if userRole == domain.RolePlanner || userRole == domain.RoleOIM || userRole == domain.RoleSystemAdmin || userRole == domain.RoleActivityOwner {
		if !targetVesselID.IsZero() {
			activities, err := s.activityRepo.FindByDateRange(ctx, targetVesselID, time.Now(), time.Now().AddDate(0, 0, 7))
			if err == nil {
				vessel, _ := s.vesselRepo.FindByID(ctx, targetVesselID)
				vesselName := ""
				if vessel != nil {
					vesselName = vessel.Name
				}
				for _, a := range activities {
					if a.Status == domain.ActivityStatusApproved || a.Status == domain.ActivityStatusSubmitted {
						data.UpcomingActivities = append(data.UpcomingActivities, UpcomingActivity{
							ID:         a.ID.Hex(),
							Name:       a.Name,
							StartDate:  a.StartDate,
							EndDate:    a.EndDate,
							Priority:   string(a.Priority),
							Status:     string(a.Status),
							VesselName: vesselName,
						})
					}
				}
			}
		}
	}

	// 3. Expiring Certificates (next 30 days)
	if userRole == domain.RoleSafetyAdmin || userRole == domain.RolePlanner || userRole == domain.RoleSystemAdmin {
		certs, err := s.certRepo.FindExpiring(ctx, 30)
		if err == nil {
			for _, c := range certs {
				personnel, _ := s.personnelRepo.FindByID(ctx, c.PersonnelID)
				name := ""
				if personnel != nil {
					name = personnel.FirstName + " " + personnel.LastName
				}
				daysRemaining := int(c.ExpiresAt.Sub(time.Now()).Hours() / 24)
				data.ExpiringCertificates = append(data.ExpiringCertificates, ExpiringCertificate{
					PersonnelID:     c.PersonnelID.Hex(),
					PersonnelName:   name,
					CertificateType: c.CertificateType,
					ExpiryDate:      c.ExpiresAt,
					DaysRemaining:   daysRemaining,
				})
			}
		}
	}

	// 4. Upcoming Travel (next 7 days)
	if userRole == domain.RolePlanner || userRole == domain.RoleSystemAdmin {
		schedules, err := s.travelSchedRepo.FindUpcoming(ctx, 100)
		if err == nil {
			for _, sch := range schedules {
				if sch.DepartureAt.After(time.Now()) && sch.DepartureAt.Before(time.Now().AddDate(0, 0, 7)) {
					transport, _ := s.transportRepo.FindByID(ctx, sch.TransportID)
					transportName := ""
					if transport != nil {
						transportName = transport.Name
					}
					vesselName := ""
					if sch.VesselID != nil {
						vessel, _ := s.vesselRepo.FindByID(ctx, *sch.VesselID)
						if vessel != nil {
							vesselName = vessel.Name
						}
					}
					data.UpcomingTravel = append(data.UpcomingTravel, UpcomingTravel{
						ScheduleID:    sch.ID.Hex(),
						TransportName: transportName,
						Direction:     string(sch.Direction),
						DepartureAt:   sch.DepartureAt,
						VesselName:    vesselName,
						SeatsBooked:   sch.ReservedSeats,
						Capacity:      sch.SeatCapacity,
					})
				}
			}
		}
	} else if userRole == domain.RolePersonnel {
		// For personnel, show only their own travels (assuming user ID equals personnel ID)
		assignments, err := s.travelAssignRepo.FindByPersonnel(ctx, userID)
		if err == nil {
			for _, assign := range assignments {
				sch, _ := s.travelSchedRepo.FindByID(ctx, assign.TravelScheduleID)
				if sch != nil && sch.DepartureAt.After(time.Now()) && sch.DepartureAt.Before(time.Now().AddDate(0, 0, 7)) {
					transport, _ := s.transportRepo.FindByID(ctx, sch.TransportID)
					transportName := ""
					if transport != nil {
						transportName = transport.Name
					}
					data.UpcomingTravel = append(data.UpcomingTravel, UpcomingTravel{
						ScheduleID:    sch.ID.Hex(),
						TransportName: transportName,
						Direction:     string(sch.Direction),
						DepartureAt:   sch.DepartureAt,
					})
				}
			}
		}
	}

	// 5. Activity Priority Distribution
	if userRole == domain.RolePlanner || userRole == domain.RoleOIM || userRole == domain.RoleSystemAdmin {
		if !targetVesselID.IsZero() {
			activities, err := s.activityRepo.FindByVessel(ctx, targetVesselID, domain.ActivityStatusApproved, domain.ActivityStatusSubmitted)
			if err == nil {
				for _, a := range activities {
					data.ActivityPriorityDist[string(a.Priority)]++
				}
			}
		}
	}

	return data, nil
}