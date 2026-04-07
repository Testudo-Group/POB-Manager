package controllers

import (
	"errors"
	"net/http"

	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{userService: userService}
}

func (c *UserController) CreateUser(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")

	var req service.CreateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user, err := c.userService.CreateUser(ctx.Request.Context(), organizationID, req)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrInvalidUserID), errors.Is(err, service.ErrInvalidUserRole):
			response.Error(ctx, http.StatusBadRequest, err.Error())
		case errors.Is(err, service.ErrUserAlreadyExists):
			response.Error(ctx, http.StatusConflict, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to create user")
		}
		return
	}

	response.Success(ctx, http.StatusCreated, "user created successfully", sanitizeUser(user))
}

func (c *UserController) ListUsers(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")
	users, err := c.userService.GetAllUsers(ctx.Request.Context(), organizationID)
	if err != nil {
		response.Error(ctx, http.StatusInternalServerError, "failed to get users")
		return
	}

	sanitizedUsers := make([]gin.H, 0, len(users))
	for _, user := range users {
		sanitizedUsers = append(sanitizedUsers, sanitizeUser(user))
	}

	response.Success(ctx, http.StatusOK, "users retrieved successfully", sanitizedUsers)
}

func (c *UserController) GetUser(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")
	id := ctx.Param("id")
	user, err := c.userService.GetUserByID(ctx.Request.Context(), organizationID, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			response.Error(ctx, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrInvalidUserID):
			response.Error(ctx, http.StatusBadRequest, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to get user")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "user retrieved successfully", sanitizeUser(user))
}

func (c *UserController) UpdateUser(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")
	id := ctx.Param("id")

	var req service.UpdateUserReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user, err := c.userService.UpdateUser(ctx.Request.Context(), organizationID, id, req)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			response.Error(ctx, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrInvalidUserID):
			response.Error(ctx, http.StatusBadRequest, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to update user")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "user updated successfully", sanitizeUser(user))
}

func (c *UserController) DeactivateUser(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")
	id := ctx.Param("id")
	err := c.userService.DeactivateUser(ctx.Request.Context(), organizationID, id)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			response.Error(ctx, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrInvalidUserID):
			response.Error(ctx, http.StatusBadRequest, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to deactivate user")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "user deactivated successfully", nil)
}

func (c *UserController) UpdateRole(ctx *gin.Context) {
	organizationID := ctx.GetString("organizationID")
	id := ctx.Param("id")

	var req service.UpdateUserRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.Error(ctx, http.StatusBadRequest, err.Error())
		return
	}

	user, err := c.userService.UpdateRole(ctx.Request.Context(), organizationID, id, req.Role)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrUserNotFound):
			response.Error(ctx, http.StatusNotFound, "user not found")
		case errors.Is(err, service.ErrInvalidUserID), errors.Is(err, service.ErrInvalidUserRole):
			response.Error(ctx, http.StatusBadRequest, err.Error())
		default:
			response.Error(ctx, http.StatusInternalServerError, "failed to update user role")
		}
		return
	}

	response.Success(ctx, http.StatusOK, "user role updated successfully", sanitizeUser(user))
}
