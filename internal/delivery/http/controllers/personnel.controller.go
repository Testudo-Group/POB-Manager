package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type PersonnelController struct {
	personnelSvc *service.PersonnelService
	compSvc      *service.ComplianceService
}

func NewPersonnelController(ps *service.PersonnelService, cs *service.ComplianceService) *PersonnelController {
	return &PersonnelController{
		personnelSvc: ps,
		compSvc:      cs,
	}
}

func (c *PersonnelController) CreatePersonnel(ctx *gin.Context) {
	var req service.CreatePersonnelInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	p, err := c.personnelSvc.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create personnel")
		return
	}

	response.Success(ctx, http.StatusCreated, "personnel created successfully", p)
}

func (c *PersonnelController) ListPersonnel(ctx *gin.Context) {
	personnelList, err := c.personnelSvc.FindAll(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get personnel")
		return
	}

	response.Success(ctx, http.StatusOK, "personnel retrieved successfully", personnelList)
}

func (c *PersonnelController) CheckCompliance(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid personnel id format")
		return
	}

	result, err := c.compSvc.CheckCompliance(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to check compliance")
		return
	}

	response.Success(ctx, http.StatusOK, "compliance status calculated", result)
}
