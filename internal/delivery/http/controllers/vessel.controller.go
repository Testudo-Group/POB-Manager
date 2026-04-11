package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VesselController struct {
	svc *service.VesselService
}

func NewVesselController(svc *service.VesselService) *VesselController {
	return &VesselController{svc: svc}
}

// CreateVessel godoc
// @Summary Create a new vessel
// @Description Creates a physical vessel or installation to track
// @Tags vessels
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param vessel body service.CreateVesselInput true "Vessel Details"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vessels [post]
func (c *VesselController) CreateVessel(ctx *gin.Context) {
	var req service.CreateVesselInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	v, err := c.svc.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create vessel")
		return
	}

	response.Success(ctx, http.StatusCreated, "vessel created successfully", v)
}

// ListVessels godoc
// @Summary List all vessels
// @Description Fetches a list of all vessels and installations
// @Tags vessels
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vessels [get]
func (c *VesselController) ListVessels(ctx *gin.Context) {
	vessels, err := c.svc.FindAll(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get vessels")
		return
	}

	response.Success(ctx, http.StatusOK, "vessels retrieved successfully", vessels)
}

func (c *VesselController) GetVessel(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	v, err := c.svc.GetByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get vessel")
		return
	}

	response.Success(ctx, http.StatusOK, "vessel retrieved successfully", v)
}

func (c *VesselController) UpdateVessel(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	var req service.CreateVesselInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	v, err := c.svc.Update(ctx.Request.Context(), id, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to update vessel")
		return
	}

	response.Success(ctx, http.StatusOK, "vessel updated successfully", v)
}

func (c *VesselController) DeleteVessel(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	err = c.svc.Delete(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to delete vessel")
		return
	}

	response.Success(ctx, http.StatusOK, "vessel deleted successfully", nil)
}

// GetRealTimePOB godoc
// @Summary Get live POB for a vessel
// @Description Fast redis-based query for real-time Personnel On Board
// @Tags vessels
// @Produce json
// @Security BearerAuth
// @Param id path string true "Vessel ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/vessels/{id}/pob [get]
func (c *VesselController) GetRealTimePOB(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	pob, err := c.svc.GetRealTimePOB(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to fetch real-time pob")
		return
	}

	response.Success(ctx, http.StatusOK, "real-time pob retrieved", gin.H{"pob": pob})
}

func (c *VesselController) GetManifest(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel id format")
		return
	}

	manifest, err := c.svc.GetManifest(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to fetch manifest")
		return
	}

	response.Success(ctx, http.StatusOK, "vessel manifest retrieved", manifest)
}
