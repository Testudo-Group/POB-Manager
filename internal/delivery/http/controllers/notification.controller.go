package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type NotificationController struct {
	svc *service.NotificationService
}

func NewNotificationController(svc *service.NotificationService) *NotificationController {
	return &NotificationController{svc: svc}
}

func (c *NotificationController) GetMyNotifications(ctx *gin.Context) {
	// Extract userID from JWT claims (set in auth middleware)
	userIDStr, exists := ctx.Get("userID")
	if !exists {
		response.Error(ctx, http.StatusUnauthorized, "user id not found in context")
		return
	}

	userID, err := bson.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "invalid user id in context")
		return
	}

	notifs, err := c.svc.GetUserNotifications(ctx.Request.Context(), userID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	response.Success(ctx, http.StatusOK, "notifications retrieved", notifs)
}

func (c *NotificationController) MarkAsRead(ctx *gin.Context) {
	idParam := ctx.Param("id")
	id, err := bson.ObjectIDFromHex(idParam)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid notification id format")
		return
	}

	err = c.svc.MarkAsRead(ctx.Request.Context(), id)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to update notification")
		return
	}

	response.Success(ctx, http.StatusOK, "notification marked as read", nil)
}
