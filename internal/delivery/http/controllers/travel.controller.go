package controllers

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type TravelController struct {
	svc *service.TravelService
}

func NewTravelController(svc *service.TravelService) *TravelController {
	return &TravelController{svc: svc}
}

// Transport Endpoints
func (c *TravelController) CreateTransport(ctx *gin.Context) {
	var req service.CreateTransportInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	transport, err := c.svc.CreateTransport(ctx.Request.Context(), req)
	if err != nil {
		c.handleTravelError(ctx, err, "failed to create transport")
		return
	}

	response.Success(ctx, http.StatusCreated, "transport created successfully", transport)
}

func (c *TravelController) ListTransports(ctx *gin.Context) {
	activeOnly := ctx.Query("active_only") == "true"
	transports, err := c.svc.ListTransports(ctx.Request.Context(), activeOnly)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to list transports")
		return
	}
	response.Success(ctx, http.StatusOK, "transports retrieved successfully", transports)
}

func (c *TravelController) GetTransport(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid transport id")
		return
	}

	transport, err := c.svc.GetTransport(ctx.Request.Context(), id)
	if err != nil {
		c.handleTravelError(ctx, err, "failed to get transport")
		return
	}

	response.Success(ctx, http.StatusOK, "transport retrieved successfully", transport)
}

func (c *TravelController) UpdateTransport(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid transport id")
		return
	}

	var req service.CreateTransportInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	transport, err := c.svc.UpdateTransport(ctx.Request.Context(), id, req)
	if err != nil {
		c.handleTravelError(ctx, err, "failed to update transport")
		return
	}

	response.Success(ctx, http.StatusOK, "transport updated successfully", transport)
}

func (c *TravelController) DeleteTransport(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid transport id")
		return
	}
	err = c.svc.DeleteTransport(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to delete transport")
		return
	}
	response.Success(ctx, http.StatusOK, "transport deleted successfully", nil)
}

// Travel Schedule Endpoints
func (c *TravelController) CreateTravelSchedule(ctx *gin.Context) {
	var req service.CreateTravelScheduleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	schedule, err := c.svc.CreateTravelSchedule(ctx.Request.Context(), req)
	if err != nil {
		c.handleTravelError(ctx, err, "failed to create travel schedule")
		return
	}

	response.Success(ctx, http.StatusCreated, "travel schedule created successfully", schedule)
}

func (c *TravelController) GetTravelSchedule(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid schedule id")
		return
	}

	schedule, err := c.svc.GetTravelSchedule(ctx.Request.Context(), id)
	if err != nil {
		c.handleTravelError(ctx, err, "failed to get travel schedule")
		return
	}

	response.Success(ctx, http.StatusOK, "travel schedule retrieved successfully", schedule)
}

func (c *TravelController) ListUpcomingSchedules(ctx *gin.Context) {
	input, err := buildTravelScheduleListInput(ctx)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	schedules, err := c.svc.ListTravelSchedules(ctx.Request.Context(), input)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to list travel schedules")
		return
	}

	response.Success(ctx, http.StatusOK, "travel schedules retrieved successfully", schedules)
}

// Auto-match
type MatchActivitiesRequest struct {
	TransportID string `json:"transport_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required"`
	EndDate     string `json:"end_date" binding:"required"`
}

func (c *TravelController) MatchActivities(ctx *gin.Context) {
	var req MatchActivitiesRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	transportID, _ := bson.ObjectIDFromHex(req.TransportID)
	startDate, _ := time.Parse(time.RFC3339, req.StartDate)
	endDate, _ := time.Parse(time.RFC3339, req.EndDate)

	activities, err := c.svc.MatchActivitiesToTransport(ctx.Request.Context(), transportID, startDate, endDate)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to match activities")
		return
	}
	response.Success(ctx, http.StatusOK, "matched activities retrieved", activities)
}

// Personnel Assignment
func (c *TravelController) AssignPersonnelToTrip(ctx *gin.Context) {
	var req service.AssignPersonnelToTripInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	err := c.svc.AssignPersonnelToTrip(ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrSeatCapacityExceeded:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}
	response.Success(ctx, http.StatusOK, "personnel assigned successfully", nil)
}

func (c *TravelController) GetTripAssignments(ctx *gin.Context) {
	scheduleID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid schedule id")
		return
	}
	assignments, err := c.svc.GetTripAssignments(ctx.Request.Context(), scheduleID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get assignments")
		return
	}
	response.Success(ctx, http.StatusOK, "assignments retrieved successfully", assignments)
}

// Utilization Alerts
func (c *TravelController) GetUtilizationAlerts(ctx *gin.Context) {
	alerts, err := c.svc.CheckLowUtilization(ctx.Request.Context(), 60.0)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to check utilization")
		return
	}
	response.Success(ctx, http.StatusOK, "utilization alerts retrieved", alerts)
}

// Trip Consolidation
type ConsolidationRequest struct {
	TransportID string `json:"transport_id" binding:"required"`
	Date        string `json:"date" binding:"required"`
}

func (c *TravelController) SuggestConsolidation(ctx *gin.Context) {
	var req ConsolidationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}
	transportID, _ := bson.ObjectIDFromHex(req.TransportID)
	date, _ := time.Parse(time.RFC3339, req.Date)

	suggestions, err := c.svc.SuggestTripConsolidation(ctx.Request.Context(), transportID, date)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get consolidation suggestions")
		return
	}
	response.Success(ctx, http.StatusOK, "consolidation suggestions retrieved", suggestions)
}

// Personnel Travel View
func (c *TravelController) GetMyTravels(ctx *gin.Context) {
	userID := ctx.GetString("user_id")
	personnelID, _ := bson.ObjectIDFromHex(userID)

	travels, err := c.svc.GetPersonnelTravels(ctx.Request.Context(), personnelID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get travels")
		return
	}
	response.Success(ctx, http.StatusOK, "travel schedule retrieved successfully", travels)
}

func buildTravelScheduleListInput(ctx *gin.Context) (service.ListTravelSchedulesInput, error) {
	var input service.ListTravelSchedulesInput

	transportID, err := optionalObjectID(ctx.Query("transport_id"))
	if err != nil {
		return input, errors.New("invalid transport_id")
	}
	input.TransportID = transportID

	vesselID, err := optionalObjectID(ctx.Query("vessel_id"))
	if err != nil {
		return input, errors.New("invalid vessel_id")
	}
	input.VesselID = vesselID

	originVesselID, err := optionalObjectID(ctx.Query("origin_vessel_id"))
	if err != nil {
		return input, errors.New("invalid origin_vessel_id")
	}
	input.OriginVesselID = originVesselID

	destinationVesselID, err := optionalObjectID(ctx.Query("destination_vessel_id"))
	if err != nil {
		return input, errors.New("invalid destination_vessel_id")
	}
	input.DestinationVesselID = destinationVesselID

	if status := ctx.Query("status"); status != "" {
		parsedStatus := domain.TravelScheduleStatus(status)
		input.Status = &parsedStatus
	}

	input.UpcomingOnly = ctx.Query("upcoming_only") == "true"
	if limit := ctx.Query("limit"); limit != "" {
		parsedLimit, err := strconv.Atoi(limit)
		if err != nil || parsedLimit < 0 {
			return input, errors.New("invalid limit")
		}
		input.Limit = parsedLimit
	}

	return input, nil
}

func optionalObjectID(raw string) (*bson.ObjectID, error) {
	if raw == "" {
		return nil, nil
	}
	id, err := bson.ObjectIDFromHex(raw)
	if err != nil {
		return nil, err
	}
	return &id, nil
}

func (c *TravelController) handleTravelError(ctx *gin.Context, err error, fallbackMessage string) {
	switch {
	case errors.Is(err, service.ErrTransportNotFound):
		response.Error(ctx, http.StatusNotFound, "transport not found")
	case errors.Is(err, service.ErrTravelScheduleNotFound):
		response.Error(ctx, http.StatusNotFound, "travel schedule not found")
	case errors.Is(err, service.ErrInvalidTravelRoute):
		response.Error(ctx, http.StatusBadRequest, err.Error())
	case errors.Is(err, repository.ErrVesselNotFound):
		response.Error(ctx, http.StatusNotFound, "vessel not found")
	default:
		response.Error(ctx, http.StatusInternalServerError, fallbackMessage)
	}
}
