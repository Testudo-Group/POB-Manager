package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type MinimumManningController struct {
	svc *service.MinimumManningService
}

func NewMinimumManningController(svc *service.MinimumManningService) *MinimumManningController {
	return &MinimumManningController{svc: svc}
}

func (c *MinimumManningController) Activate(ctx *gin.Context) {
	var req service.ActivateMinimumManningInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	event, err := c.svc.Activate(ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrMinimumManningAlreadyActive:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to activate minimum manning mode")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "minimum manning mode activated successfully", event)
}

func (c *MinimumManningController) Deactivate(ctx *gin.Context) {
	var req service.DeactivateMinimumManningInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	event, err := c.svc.Deactivate(ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrMinimumManningNotActive:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to deactivate minimum manning mode")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "minimum manning mode deactivated successfully", event)
}

func (c *MinimumManningController) GetActiveEvent(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	event, err := c.svc.GetActiveEvent(ctx.Request.Context(), vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get active event")
		return
	}

	response.Success(ctx, http.StatusOK, "active minimum manning event retrieved", event)
}

func (c *MinimumManningController) GetEventHistory(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	events, err := c.svc.GetEventHistory(ctx.Request.Context(), vesselID, 50)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get event history")
		return
	}

	response.Success(ctx, http.StatusOK, "minimum manning event history retrieved", events)
}