package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/codingninja/pob-management/internal/repository"
	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReportService struct {
	vesselRepo     *repository.VesselRepository
	activityRepo   *repository.ActivityRepository
	personnelRepo  *repository.PersonnelRepository
	roomAssignRepo *repository.RoomAssignmentRepository
	vesselSvc      *VesselService
}

func NewReportService(
	vesselRepo *repository.VesselRepository,
	activityRepo *repository.ActivityRepository,
	personnelRepo *repository.PersonnelRepository,
	roomAssignRepo *repository.RoomAssignmentRepository,
	vesselSvc *VesselService,
) *ReportService {
	return &ReportService{
		vesselRepo:     vesselRepo,
		activityRepo:   activityRepo,
		personnelRepo:  personnelRepo,
		roomAssignRepo: roomAssignRepo,
		vesselSvc:      vesselSvc,
	}
}

type DailyPOBReport struct {
	Date        time.Time          `json:"date"`
	VesselID    string             `json:"vessel_id"`
	VesselName  string             `json:"vessel_name"`
	POB         int                `json:"pob"`
	Capacity    int                `json:"capacity"`
	Utilization float64            `json:"utilization"`
	Manifest    []PersonnelOnBoard `json:"manifest"`
}

type PersonnelOnBoard struct {
	PersonnelID  string    `json:"personnel_id"`
	Name         string    `json:"name"`
	Role         string    `json:"role"`
	Room         string    `json:"room"`
	OnboardSince time.Time `json:"onboard_since"`
}

func (s *ReportService) GenerateDailyPOB(ctx context.Context, vesselID bson.ObjectID, date time.Time) (*DailyPOBReport, error) {
	vessel, err := s.vesselRepo.FindByID(ctx, vesselID)
	if err != nil {
		return nil, err
	}

	pob, _ := s.vesselSvc.GetRealTimePOB(ctx, vesselID)
	assignments, err := s.roomAssignRepo.FindActiveByVessel(ctx, vesselID)
	if err != nil {
		return nil, err
	}

	manifest := []PersonnelOnBoard{}
	for _, assign := range assignments {
		personnel, _ := s.personnelRepo.FindByID(ctx, assign.PersonnelID)
		if personnel != nil {
			manifest = append(manifest, PersonnelOnBoard{
				PersonnelID:  personnel.ID.Hex(),
				Name:         personnel.FirstName + " " + personnel.LastName,
				Role:         "", // Could be enhanced with role fetch
				Room:         assign.RoomID.Hex(),
				OnboardSince: assign.StartsAt,
			})
		}
	}

	utilization := 0.0
	if vessel.POBCapacity > 0 {
		utilization = float64(pob) / float64(vessel.POBCapacity) * 100
	}

	return &DailyPOBReport{
		Date:        date,
		VesselID:    vesselID.Hex(),
		VesselName:  vessel.Name,
		POB:         pob,
		Capacity:    vessel.POBCapacity,
		Utilization: utilization,
		Manifest:    manifest,
	}, nil
}

func (s *ReportService) GenerateHistoricalPOB(ctx context.Context, vesselID bson.ObjectID, start, end time.Time) ([]DailyPOBReport, error) {
	var reports []DailyPOBReport
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		report, err := s.GenerateDailyPOB(ctx, vesselID, d)
		if err == nil {
			reports = append(reports, *report)
		}
	}
	return reports, nil
}

func (s *ReportService) ExportDailyPOBPDF(ctx context.Context, vesselID bson.ObjectID, date time.Time) ([]byte, error) {
	report, err := s.GenerateDailyPOB(ctx, vesselID, date)
	if err != nil {
		return nil, err
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(40, 10, fmt.Sprintf("Daily POB Report - %s", report.Date.Format("2006-01-02")))
	pdf.Ln(12)
	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Vessel: %s", report.VesselName))
	pdf.Ln(8)
	pdf.Cell(40, 10, fmt.Sprintf("POB: %d / %d (%.1f%%)", report.POB, report.Capacity, report.Utilization))
	pdf.Ln(12)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Manifest:")
	pdf.Ln(8)
	pdf.SetFont("Arial", "", 10)
	for _, p := range report.Manifest {
		pdf.Cell(0, 6, fmt.Sprintf("%s - Room %s", p.Name, p.Room))
		pdf.Ln(6)
	}

	var buf bytes.Buffer
	err = pdf.Output(&buf)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (s *ReportService) ExportDailyPOBCSV(ctx context.Context, vesselID bson.ObjectID, date time.Time) ([]byte, error) {
	report, err := s.GenerateDailyPOB(ctx, vesselID, date)
	if err != nil {
		return nil, err
	}

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Sheet1"

	f.SetCellValue(sheet, "A1", "Date")
	f.SetCellValue(sheet, "B1", report.Date.Format("2006-01-02"))
	f.SetCellValue(sheet, "A2", "Vessel")
	f.SetCellValue(sheet, "B2", report.VesselName)
	f.SetCellValue(sheet, "A3", "POB")
	f.SetCellValue(sheet, "B3", report.POB)
	f.SetCellValue(sheet, "A4", "Capacity")
	f.SetCellValue(sheet, "B4", report.Capacity)
	f.SetCellValue(sheet, "A5", "Utilization %")
	f.SetCellValue(sheet, "B5", report.Utilization)

	f.SetCellValue(sheet, "A7", "Personnel")
	f.SetCellValue(sheet, "B7", "Room")
	f.SetCellValue(sheet, "C7", "Onboard Since")
	for i, p := range report.Manifest {
		row := i + 8
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), p.Name)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), p.Room)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), p.OnboardSince.Format("2006-01-02"))
	}

	buf, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}