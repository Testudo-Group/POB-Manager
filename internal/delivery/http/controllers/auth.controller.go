package controllers

import (
	"errors"
	"net/http"

	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService *service.AuthService
}

type registerRequest struct {
	FirstName string          `json:"first_name" binding:"required"`
	LastName  string          `json:"last_name" binding:"required"`
	Email     string          `json:"email" binding:"required,email"`
	Password  string          `json:"password" binding:"required,min=8"`
	Role      domain.UserRole `json:"role" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type updateMeRequest struct {
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

func NewAuthController(authService *service.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctl *AuthController) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	if !isValidRole(req.Role) {
		response.Error(c, http.StatusBadRequest, "invalid role")
		return
	}

	user, tokens, err := ctl.authService.Register(c.Request.Context(), service.RegisterInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Password:  req.Password,
		Role:      req.Role,
	})
	if err != nil {
		if errors.Is(err, service.ErrEmailAlreadyExists) {
			response.Error(c, http.StatusConflict, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to register user")
		return
	}

	response.Success(c, http.StatusCreated, "user registered successfully", gin.H{
		"user":   sanitizeUser(user),
		"tokens": tokens,
	})
}

func (ctl *AuthController) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, tokens, err := ctl.authService.Login(c.Request.Context(), service.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to login")
		return
	}

	response.Success(c, http.StatusOK, "login successful", gin.H{
		"user":   sanitizeUser(user),
		"tokens": tokens,
	})
}

func (ctl *AuthController) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	user, tokens, err := ctl.authService.Refresh(c.Request.Context(), req.RefreshToken)
	if err != nil {
		if errors.Is(err, service.ErrInvalidRefreshToken) {
			response.Error(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to refresh token")
		return
	}

	response.Success(c, http.StatusOK, "token refreshed successfully", gin.H{
		"user":   sanitizeUser(user),
		"tokens": tokens,
	})
}

func (ctl *AuthController) Logout(c *gin.Context) {
	userID := c.GetString("userID")
	if err := ctl.authService.Logout(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to logout")
		return
	}

	response.Success(c, http.StatusOK, "logout successful", nil)
}

func (ctl *AuthController) Me(c *gin.Context) {
	userID := c.GetString("userID")
	user, err := ctl.authService.Me(c.Request.Context(), userID)
	if err != nil {
		response.Error(c, http.StatusNotFound, "user not found")
		return
	}

	response.Success(c, http.StatusOK, "current user fetched successfully", sanitizeUser(user))
}

func (ctl *AuthController) UpdateMe(c *gin.Context) {
	var req updateMeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetString("userID")
	user, err := ctl.authService.UpdateMe(c.Request.Context(), userID, service.UpdateProfileInput{
		FirstName: req.FirstName,
		LastName:  req.LastName,
	})
	if err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to update profile")
		return
	}

	response.Success(c, http.StatusOK, "profile updated successfully", sanitizeUser(user))
}

func (ctl *AuthController) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userID := c.GetString("userID")
	err := ctl.authService.ChangePassword(c.Request.Context(), userID, service.ChangePasswordInput{
		CurrentPassword: req.CurrentPassword,
		NewPassword:     req.NewPassword,
	})
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			response.Error(c, http.StatusUnauthorized, "current password is incorrect")
			return
		}
		response.Error(c, http.StatusInternalServerError, "failed to change password")
		return
	}

	response.Success(c, http.StatusOK, "password changed successfully", nil)
}

func sanitizeUser(user *domain.User) gin.H {
	return gin.H{
		"id":         user.ID.Hex(),
		"first_name": user.FirstName,
		"last_name":  user.LastName,
		"email":      user.Email,
		"role":       user.Role,
		"is_active":  user.IsActive,
		"created_at": user.CreatedAt,
		"updated_at": user.UpdatedAt,
	}
}

func isValidRole(role domain.UserRole) bool {
	switch role {
	case domain.RoleActivityOwner, domain.RolePlanner, domain.RoleSafetyAdmin, domain.RoleOIM, domain.RolePersonnel, domain.RoleSysAdmin:
		return true
	default:
		return false
	}
}
