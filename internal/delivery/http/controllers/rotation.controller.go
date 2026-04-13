package controllers

import (
	"net/http"
	"time"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RotationController struct {
	svc *service.RotationService
}

func NewRotationController(svc *service.RotationService) *RotationController {
	return &RotationController{svc: svc}
}

// Rotation Schedules
func (c *RotationController) CreateSchedule(ctx *gin.Context) {
	var req service.CreateRotationScheduleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	schedule, err := c.svc.CreateSchedule(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create rotation schedule")
		return
	}

	response.Success(ctx, http.StatusCreated, "rotation schedule created successfully", schedule)
}

func (c *RotationController) GetSchedules(ctx *gin.Context) {
	roleID, err := bson.ObjectIDFromHex(ctx.Query("role_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role_id")
		return
	}

	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	schedules, err := c.svc.GetSchedulesByRole(ctx.Request.Context(), roleID, vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get rotation schedules")
		return
	}

	response.Success(ctx, http.StatusOK, "rotation schedules retrieved successfully", schedules)
}

// Role Assignments
func (c *RotationController) AssignRole(ctx *gin.Context) {
	var req service.AssignRoleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	assignment, err := c.svc.AssignRole(ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrRoleAlreadyAssigned:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to assign role")
		}
		return
	}

	response.Success(ctx, http.StatusCreated, "role assigned successfully", assignment)
}

type EndAssignmentRequest struct {
	EffectiveTo string `json:"effective_to" binding:"required"`
}

func (c *RotationController) EndAssignment(ctx *gin.Context) {
	assignmentID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid assignment id")
		return
	}

	var req EndAssignmentRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	effectiveTo, err := time.Parse(time.RFC3339, req.EffectiveTo)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid effective_to date format")
		return
	}

	err = c.svc.EndAssignment(ctx.Request.Context(), assignmentID, effectiveTo)
	if err != nil {
		switch err {
		case service.ErrMinimumRoleCount:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to end assignment")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "assignment ended successfully", nil)
}

func (c *RotationController) GetVesselManning(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Param("vesselId"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id")
		return
	}

	assignments, err := c.svc.GetVesselManning(ctx.Request.Context(), vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get vessel manning")
		return
	}

	response.Success(ctx, http.StatusOK, "vessel manning retrieved successfully", assignments)
}

// Back-to-Back Pairs
func (c *RotationController) CreateBackToBackPair(ctx *gin.Context) {
	var req service.CreateBackToBackPairInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	pair, err := c.svc.CreateBackToBackPair(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create back-to-back pair")
		return
	}

	response.Success(ctx, http.StatusCreated, "back-to-back pair created successfully", pair)
}

func (c *RotationController) GetBackToBackPairs(ctx *gin.Context) {
	roleID, err := bson.ObjectIDFromHex(ctx.Query("role_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role_id")
		return
	}

	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	pairs, err := c.svc.GetBackToBackPairsByRole(ctx.Request.Context(), roleID, vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get back-to-back pairs")
		return
	}

	response.Success(ctx, http.StatusOK, "back-to-back pairs retrieved successfully", pairs)
}

// Calculate Next Rotation
type CalculateRotationRequest struct {
	ScheduleID string `json:"schedule_id" binding:"required"`
	FromDate   string `json:"from_date" binding:"required"`
}

func (c *RotationController) CalculateNextRotation(ctx *gin.Context) {
	var req CalculateRotationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	scheduleID, err := bson.ObjectIDFromHex(req.ScheduleID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid schedule_id")
		return
	}

	fromDate, err := time.Parse(time.RFC3339, req.FromDate)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid from_date format")
		return
	}

	nextOn, nextOff, err := c.svc.CalculateNextRotation(ctx.Request.Context(), scheduleID, fromDate)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to calculate rotation")
		return
	}

	response.Success(ctx, http.StatusOK, "rotation calculated successfully", gin.H{
		"next_on_date":  nextOn,
		"next_off_date": nextOff,
	})
}