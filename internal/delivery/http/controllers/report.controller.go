package controllers

import (
	"net/http"
	"time"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ReportController struct {
	reportSvc *service.ReportService
}

func NewReportController(svc *service.ReportService) *ReportController {
	return &ReportController{reportSvc: svc}
}

func (c *ReportController) DailyPOBReport(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}
	dateStr := ctx.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	report, err := c.reportSvc.GenerateDailyPOB(ctx.Request.Context(), vesselID, date)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to generate report")
		return
	}
	response.Success(ctx, http.StatusOK, "daily POB report", report)
}

func (c *ReportController) HistoricalPOBReport(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}
	start, _ := time.Parse("2006-01-02", ctx.Query("start_date"))
	end, _ := time.Parse("2006-01-02", ctx.Query("end_date"))
	if start.IsZero() {
		start = time.Now().AddDate(0, 0, -7)
	}
	if end.IsZero() {
		end = time.Now()
	}

	reports, err := c.reportSvc.GenerateHistoricalPOB(ctx.Request.Context(), vesselID, start, end)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to generate historical report")
		return
	}
	response.Success(ctx, http.StatusOK, "historical POB report", reports)
}

func (c *ReportController) ExportPDF(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}
	dateStr := ctx.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	pdfData, err := c.reportSvc.ExportDailyPOBPDF(ctx.Request.Context(), vesselID, date)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to generate PDF")
		return
	}

	ctx.Header("Content-Type", "application/pdf")
	ctx.Header("Content-Disposition", "attachment; filename=daily_pob_report.pdf")
	ctx.Data(http.StatusOK, "application/pdf", pdfData)
}

func (c *ReportController) ExportCSV(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}
	dateStr := ctx.Query("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		date = time.Now()
	}

	csvData, err := c.reportSvc.ExportDailyPOBCSV(ctx.Request.Context(), vesselID, date)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to generate CSV")
		return
	}

	ctx.Header("Content-Type", "text/csv")
	ctx.Header("Content-Disposition", "attachment; filename=daily_pob_report.csv")
	ctx.Data(http.StatusOK, "text/csv", csvData)
}