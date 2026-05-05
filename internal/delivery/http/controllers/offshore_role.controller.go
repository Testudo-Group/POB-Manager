package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type OffshoreRoleController struct {
	svc *service.OffshoreRoleService
}

func NewOffshoreRoleController(svc *service.OffshoreRoleService) *OffshoreRoleController {
	return &OffshoreRoleController{svc: svc}
}

func (c *OffshoreRoleController) CreateRole(ctx *gin.Context) {
	var req service.CreateOffshoreRoleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	role, err := c.svc.Create(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to create role")
		return
	}

	response.Success(ctx, http.StatusCreated, "role created successfully", role)
}

func (c *OffshoreRoleController) ListRoles(ctx *gin.Context) {
	roles, err := c.svc.FindAll(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get roles")
		return
	}

	response.Success(ctx, http.StatusOK, "roles retrieved successfully", roles)
}

func (c *OffshoreRoleController) GetRole(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role id")
		return
	}

	role, err := c.svc.FindByID(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "role not found")
		return
	}

	response.Success(ctx, http.StatusOK, "role retrieved successfully", role)
}

func (c *OffshoreRoleController) UpdateRole(ctx *gin.Context) {
	id, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role id")
		return
	}

	var req service.CreateOffshoreRoleInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	role, err := c.svc.Update(ctx.Request.Context(), id, req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to update role")
		return
	}

	response.Success(ctx, http.StatusOK, "role updated successfully", role)
}

type CertTypeBody struct {
	CertificateTypeID string `json:"certificate_type_id" binding:"required"`
}

func (c *OffshoreRoleController) AddRequiredCertificate(ctx *gin.Context) {
	roleID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role id")
		return
	}

	var body CertTypeBody
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	if body.CertificateTypeID == "" {
		response.Error(ctx, http.StatusBadRequest, "certificate_type_id is required")
		return
	}

	role, err := c.svc.AddRequiredCertificate(ctx.Request.Context(), roleID, body.CertificateTypeID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to add required certificate")
		return
	}

	response.Success(ctx, http.StatusOK, "required certificate added", role)
}

func (c *OffshoreRoleController) RemoveRequiredCertificate(ctx *gin.Context) {
	roleID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid role id")
		return
	}

	certTypeCode := ctx.Param("certTypeId")
	if certTypeCode == "" {
		response.Error(ctx, http.StatusBadRequest, "certificate type code is required")
		return
	}

	role, err := c.svc.RemoveRequiredCertificate(ctx.Request.Context(), roleID, certTypeCode)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to remove required certificate")
		return
	}

	response.Success(ctx, http.StatusOK, "required certificate removed", role)
}
