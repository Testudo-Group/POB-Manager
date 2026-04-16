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
	OrganizationName    string `json:"organization_name" binding:"required"`
	OrganizationPhone   string `json:"organization_phone" binding:"required"`
	OrganizationAddress string `json:"organization_address" binding:"required"`
	FirstName           string `json:"first_name" binding:"required"`
	LastName            string `json:"last_name" binding:"required"`
	PhoneNumber         string `json:"phone_number" binding:"required"`
	Email               string `json:"email" binding:"required,email"`
	Password            string `json:"password" binding:"required,min=8"`
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

	user, tokens, err := ctl.authService.Register(c.Request.Context(), service.RegisterInput{
		OrganizationName:    req.OrganizationName,
		OrganizationPhone:   req.OrganizationPhone,
		OrganizationAddress: req.OrganizationAddress,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		PhoneNumber:         req.PhoneNumber,
		Email:               req.Email,
		Password:            req.Password,
	})
	if err != nil {
		switch {
		case errors.Is(err, service.ErrEmailAlreadyExists), errors.Is(err, service.ErrOrganizationExists):
			response.Error(c, http.StatusConflict, err.Error())
		default:
			response.Error(c, http.StatusInternalServerError, "failed to register organization")
		}
		return
	}

	response.Success(c, http.StatusCreated, "organization registered successfully", gin.H{
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
	userID := c.GetString("user_id")
	if err := ctl.authService.Logout(c.Request.Context(), userID); err != nil {
		response.Error(c, http.StatusInternalServerError, "failed to logout")
		return
	}

	response.Success(c, http.StatusOK, "logout successful", nil)
}

func (ctl *AuthController) Me(c *gin.Context) {
	userID := c.GetString("user_id")
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

	userID := c.GetString("user_id")
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

	userID := c.GetString("user_id")
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
	data := gin.H{
		"id":              user.ID.Hex(),
		"organization_id": user.OrganizationID.Hex(),
		"first_name":      user.FirstName,
		"last_name":       user.LastName,
		"phone_number":    user.PhoneNumber,
		"email":           user.Email,
		"role":            user.Role,
		"is_active":       user.IsActive,
		"created_at":      user.CreatedAt,
		"updated_at":      user.UpdatedAt,
	}
	if user.VesselID != nil {
		data["vessel_id"] = user.VesselID.Hex()
	}
	return data
}
