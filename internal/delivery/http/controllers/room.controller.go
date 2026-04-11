package controllers

import (
	"errors"
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type RoomController struct {
	svc *service.RoomService
}

func NewRoomController(svc *service.RoomService) *RoomController {
	return &RoomController{svc: svc}
}

// CreateRoom godoc
// @Summary Create a room inside a vessel
// @Description Sets up a physical room with beds
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vesselId path string true "Vessel ID"
// @Param room body service.CreateRoomInput true "Room Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vessels/{vesselId}/rooms [post]
func (c *RoomController) CreateRoom(ctx *gin.Context) {
	var req service.CreateRoomInput
	
	// Inject vessel ID from path
	idParam := ctx.Param("vesselId")
	vesselID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}
	req.VesselID = vesselID

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	r, err := c.svc.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create room")
		return
	}

	response.Success(ctx, http.StatusCreated, "room created successfully", r)
}

// AssignRoom godoc
// @Summary Assign personnel to a room bed
// @Description Assigns someone to a room and inherently increments the real-time POB
// @Tags rooms
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vesselId path string true "Vessel ID"
// @Param room body service.AssignRoomInput true "Assignment logic"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vessels/{vesselId}/rooms/assign [post]
func (c *RoomController) AssignRoom(ctx *gin.Context) {
	var req service.AssignRoomInput
	
	idParam := ctx.Param("vesselId")
	vesselID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}
	req.VesselID = vesselID

	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	assignment, err := c.svc.Assign(ctx.Request.Context(), req)
	if err != nil {
		if errors.Is(err, service.ErrRoomCapacityExceeded) || errors.Is(err, service.ErrVesselCapacityBlocked) {
			response.Error(ctx, http.StatusBadRequest, err.Error())
			return
		}
		response.Error(ctx, http.StatusInternalServerError, "failed to assign room")
		return
	}

	response.Success(ctx, http.StatusCreated, "room successfully assigned", assignment)
}
