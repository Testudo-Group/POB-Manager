package middleware

import (
	"net/http"
	"strings"

	"github.com/codingninja/pob-management/internal/service"
	"github.com/codingninja/pob-management/pkg/response"
	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	tokenManager *service.TokenManager
}

func NewAuthMiddleware(tokenManager *service.TokenManager) *AuthMiddleware {
	return &AuthMiddleware{tokenManager: tokenManager}
}

func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "authorization header is required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			response.Error(c, http.StatusUnauthorized, "authorization header must use Bearer token")
			c.Abort()
			return
		}

		claims, err := m.tokenManager.Validate(parts[1], service.AccessTokenType)
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "invalid or expired access token")
			c.Abort()
			return
		}

		c.Set("user_id", claims.Subject)
		c.Set("organization_id", claims.OrganizationID)
		c.Set("user_role", claims.Role)
		c.Set("user_email", claims.Email)
		c.Next()
	}
}
