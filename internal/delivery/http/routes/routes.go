package routes

import (
	"context"
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/controllers"
	"github.com/codingninja/pob-management/internal/delivery/http/middleware"
	"github.com/codingninja/pob-management/internal/repository"
	"github.com/codingninja/pob-management/internal/service"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func Setup(r *gin.Engine, db *mongo.Database, cfg *config.Config) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	userRepository := repository.NewUserRepository(db)
	tokenManager := service.NewTokenManager(cfg.JWTSecret, cfg.AccessTokenTTLMinutes, cfg.RefreshTokenTTLHours)
	authService := service.NewAuthService(userRepository, tokenManager)
	authController := controllers.NewAuthController(authService)
	authMiddleware := middleware.NewAuthMiddleware(tokenManager)

	if err := authService.Initialize(context.Background()); err != nil {
		log.Printf("failed to initialize auth indexes: %v", err)
	}

	api := r.Group("/api/v1")
	auth := api.Group("/auth")
	{
		auth.POST("/register", authController.Register)
		auth.POST("/login", authController.Login)
		auth.POST("/refresh", authController.Refresh)

		authenticated := auth.Group("")
		authenticated.Use(authMiddleware.RequireAuth())
		{
			authenticated.POST("/logout", authController.Logout)
			authenticated.GET("/me", authController.Me)
			authenticated.PATCH("/me", authController.UpdateMe)
			authenticated.POST("/change-password", authController.ChangePassword)
		}
	}
}
