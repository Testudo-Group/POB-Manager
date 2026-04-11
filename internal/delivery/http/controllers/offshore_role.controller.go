package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
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
