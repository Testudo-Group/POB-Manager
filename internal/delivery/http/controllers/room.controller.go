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

func (c *RoomController) ListRooms(ctx *gin.Context) {
	idParam := ctx.Param("vesselId")
	vesselID, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	rooms, err := c.svc.ListByVessel(ctx.Request.Context(), vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get rooms")
		return
	}

	response.Success(ctx, http.StatusOK, "rooms retrieved successfully", rooms)
}

func (c *RoomController) GetRoom(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid room id format")
		return
	}

	r, err := c.svc.FindByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get room")
		return
	}

	response.Success(ctx, http.StatusOK, "room retrieved successfully", r)
}

func (c *RoomController) UpdateRoom(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid room id format")
		return
	}

	var req service.CreateRoomInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	r, err := c.svc.Update(ctx.Request.Context(), id, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to update room")
		return
	}

	response.Success(ctx, http.StatusOK, "room updated successfully", r)
}

func (c *RoomController) DeleteRoom(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid room id format")
		return
	}

	err = c.svc.Delete(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to delete room")
		return
	}

	response.Success(ctx, http.StatusOK, "room deleted successfully", nil)
}

func (c *RoomController) GetRoomOccupants(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid room id format")
		return
	}

	occupants, err := c.svc.GetOccupants(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get room occupants")
		return
	}

	response.Success(ctx, http.StatusOK, "room occupants retrieved successfully", occupants)
}
