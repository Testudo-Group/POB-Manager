package main

import (
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/routes"
	"github.com/codingninja/pob-management/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load config
	cfg := config.Load()

	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Setup Gin
	r := gin.Default()

	// Register routes
	routes.Setup(r, db, cfg)

	// Start server
	log.Println("Server running on port", cfg.Port)
	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
