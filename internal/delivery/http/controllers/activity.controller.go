package controllers

import (
	"net/http"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ActivityController struct {
	svc *service.ActivityService
}

func NewActivityController(svc *service.ActivityService) *ActivityController {
	return &ActivityController{svc: svc}
}

func (c *ActivityController) CreateActivity(ctx *gin.Context) {
	var req service.CreateActivityInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	activity, err := c.svc.Create(ctx.Request.Context(), req)
	if err != nil {
		switch err {
		case service.ErrActivityConflict:
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, err.Error())
		}
		return
	}

	response.Success(ctx, http.StatusCreated, "activity created successfully", activity)
}

func (c *ActivityController) ListActivities(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	activities, err := c.svc.GetByVessel(ctx.Request.Context(), vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get activities")
		return
	}

	response.Success(ctx, http.StatusOK, "activities retrieved successfully", activities)
}

type SubmitRequest struct {
	ActivityID string `json:"activity_id" binding:"required"`
}

func (c *ActivityController) SubmitForApproval(ctx *gin.Context) {
	var req SubmitRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	activityID, err := bson.ObjectIDFromHex(req.ActivityID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity_id")
		return
	}

	userID := ctx.GetString("user_id")
	userObjID, _ := bson.ObjectIDFromHex(userID)

	err = c.svc.SubmitForApproval(ctx.Request.Context(), activityID, userObjID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "activity submitted for approval", nil)
}

func (c *ActivityController) ApproveActivity(ctx *gin.Context) {
	var req service.ApproveActivityInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	err := c.svc.Approve(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "activity approved successfully", nil)
}

type RejectRequest struct {
	ActivityID string `json:"activity_id" binding:"required"`
	Note       string `json:"note"`
}

func (c *ActivityController) RejectActivity(ctx *gin.Context) {
	var req RejectRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	activityID, err := bson.ObjectIDFromHex(req.ActivityID)
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity_id")
		return
	}

	userID := ctx.GetString("user_id")
	userObjID, _ := bson.ObjectIDFromHex(userID)

	err = c.svc.Reject(ctx.Request.Context(), activityID, userObjID, req.Note)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "activity rejected", nil)
}

func (c *ActivityController) GetPendingQueue(ctx *gin.Context) {
	activities, err := c.svc.GetPendingApproval(ctx.Request.Context())
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get pending activities")
		return
	}

	response.Success(ctx, http.StatusOK, "pending activities retrieved successfully", activities)
}

func (c *ActivityController) GetGanttData(ctx *gin.Context) {
	vesselID, err := bson.ObjectIDFromHex(ctx.Query("vessel_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid vessel_id")
		return
	}

	ganttData, err := c.svc.GetGanttData(ctx.Request.Context(), vesselID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get gantt data")
		return
	}

	response.Success(ctx, http.StatusOK, "gantt data retrieved successfully", ganttData)
}

func (c *ActivityController) CheckConflicts(ctx *gin.Context) {
	activityID, err := bson.ObjectIDFromHex(ctx.Query("activity_id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity_id")
		return
	}

	conflicts, err := c.svc.CheckConflicts(ctx.Request.Context(), activityID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to check conflicts")
		return
	}

	response.Success(ctx, http.StatusOK, "conflict check completed", gin.H{
		"has_conflicts": len(conflicts) > 0,
		"conflicts":     conflicts,
	})
}

func (c *ActivityController) GetRequirements(ctx *gin.Context) {
	activityID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity id")
		return
	}

	reqs, err := c.svc.GetRequirements(ctx.Request.Context(), activityID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get requirements")
		return
	}

	response.Success(ctx, http.StatusOK, "requirements retrieved successfully", reqs)
}

func (c *ActivityController) AssignPersonnel(ctx *gin.Context) {
	var req service.AssignPersonnelInput
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	err := c.svc.AssignPersonnel(ctx.Request.Context(), req)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "personnel assigned successfully", nil)
}

func (c *ActivityController) GetAssignments(ctx *gin.Context) {
	activityID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity id")
		return
	}

	assignments, err := c.svc.GetAssignments(ctx.Request.Context(), activityID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get assignments")
		return
	}

	response.Success(ctx, http.StatusOK, "assignments retrieved successfully", assignments)
}

func (c *ActivityController) DeleteActivity(ctx *gin.Context) {
	activityID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity id")
		return
	}

	err = c.svc.Delete(ctx.Request.Context(), activityID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, err.Error())
		return
	}

	response.Success(ctx, http.StatusOK, "activity deleted successfully", nil)
}
func (c *ActivityController) GetActivity(ctx *gin.Context) {
	activityID, err := bson.ObjectIDFromHex(ctx.Param("id"))
	if err != nil {
		response.Error(ctx, http.StatusBadRequest, "invalid activity id")
		return
	}

	activity, err := c.svc.GetByID(ctx.Request.Context(), activityID)
	if err != nil {
		response.Error(ctx, http.StatusNotFound, "activity not found")
		return
	}

	response.Success(ctx, http.StatusOK, "activity retrieved successfully", activity)
}
