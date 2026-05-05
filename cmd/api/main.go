package main

import (
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/middleware"
	"github.com/codingninja/pob-management/internal/delivery/http/routes"
	"github.com/codingninja/pob-management/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {

	cfg := config.Load()

	// MongoDB
	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	log.Println("✅ MongoDB connected successfully")

	// Upstash Redis (REST) - returns standard *redis.Client
	rdb := database.ConnectRedis(
		cfg.UpstashRedisURL,
		cfg.UpstashRedisToken,
	)

	// Gin
	r := gin.Default()

	r.Use(middleware.CORSMiddleware())

	routes.Setup(r, db, rdb, cfg)

	log.Println("Server running on port", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}