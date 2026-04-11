package main

import (
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/routes"
	"github.com/codingninja/pob-management/pkg/database"
	"github.com/gin-gonic/gin"
)

// @title POB Management API
// @version 1.0
// @description API for managing Personnel On Board, Vessels, and Rooms.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization

// @host localhost:8080
// @BasePath /
func main() {
	// Load config
	cfg := config.Load()

	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	rdb, err := database.ConnectRedis(cfg.RedisURI)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	// Setup Gin
	r := gin.Default()

	// Register routes
	routes.Setup(r, db, rdb, cfg)

	// Start server
	log.Println("Server running on port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
