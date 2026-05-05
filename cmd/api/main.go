package main

import (
	"context"
	"log"

	"github.com/codingninja/pob-management/config"
	"github.com/codingninja/pob-management/internal/delivery/http/middleware"
	"github.com/codingninja/pob-management/internal/delivery/http/routes"
	"github.com/codingninja/pob-management/pkg/database"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration (will fail if Redis env vars are missing)
	log.Println("🚀 Starting POB Management Server...")
	cfg := config.Load()

	// Connect to MongoDB
	log.Println("📦 Connecting to MongoDB...")
	db, err := database.NewMongoDatabase(cfg)
	if err != nil {
		log.Fatal("❌ Failed to connect to database:", err)
	}
	log.Println("✅ MongoDB connected successfully")

	// Connect to Upstash Redis (REST API)
	log.Println("🗄️ Connecting to Upstash Redis...")
	rdb := database.ConnectRedis(
		cfg.UpstashRedisURL,
		cfg.UpstashRedisToken,
	)

	// Test Redis connection
	ctx := context.Background()
	if err := rdb.Ping(ctx); err != nil {
		log.Fatalf("❌ Redis connection failed: %v", err)
	}
	log.Println("✅ Redis connected successfully")

	// Initialize Gin router
	r := gin.Default()

	// Add middleware
	r.Use(middleware.CORSMiddleware())

	// Setup routes
	routes.Setup(r, db, rdb, cfg)

	// Start server
	log.Printf("🌐 Server running on port %s", cfg.Port)
	log.Printf("📝 Health check: http://localhost:%s/health", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
