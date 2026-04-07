package middleware

import (
	"net/http"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/domain"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
)

func RequireRole(allowedRoles ...domain.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user role not found in context")
			c.Abort()
			return
		}

		var userRole domain.UserRole
		switch value := roleVal.(type) {
		case domain.UserRole:
			userRole = value
		case string:
			userRole = domain.UserRole(value)
		default:
			response.Error(c, http.StatusInternalServerError, "invalid role format in context")
			c.Abort()
			return
		}

		allowed := false
		for _, role := range allowedRoles {
			if userRole == role {
				allowed = true
				break
			}
		}

		if !allowed {
			response.Error(c, http.StatusForbidden, "you do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}

func RequirePermission(permission config.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("userRole")
		if !exists {
			response.Error(c, http.StatusUnauthorized, "user role not found in context")
			c.Abort()
			return
		}

		roleString, ok := roleVal.(string)
		if !ok {
			response.Error(c, http.StatusInternalServerError, "invalid role format in context")
			c.Abort()
			return
		}

		if !config.HasPermission(config.Role(roleString), permission) {
			response.Error(c, http.StatusForbidden, "you do not have permission to perform this action")
			c.Abort()
			return
		}

		c.Next()
	}
}
