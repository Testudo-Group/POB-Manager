package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type DashboardController struct {
	dashboardSvc *service.DashboardService
}

func NewDashboardController(svc *service.DashboardService) *DashboardController {
	return &DashboardController{dashboardSvc: svc}
}

func (c *DashboardController) GetDashboard(ctx *gin.Context) {
	userIDStr := ctx.GetString("user_id")
	userID, err := bson.ObjectIDFromHex(userIDStr)
	if err != nil {
		response.Error(ctx, http.StatusUnauthorized, "invalid user id")
		return
	}

	userRoleStr := ctx.GetString("user_role")
	userRole := domain.UserRole(userRoleStr)

	var vesselID *bson.ObjectID
	if vid := ctx.Query("vessel_id"); vid != "" {
		id, err := bson.ObjectIDFromHex(vid)
		if err == nil {
			vesselID = &id
		}
	}

	data, err := c.dashboardSvc.GetRoleFilteredDashboard(ctx.Request.Context(), userID, userRole, vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to load dashboard")
		return
	}

	response.Success(ctx, http.StatusOK, "dashboard data retrieved", data)
}