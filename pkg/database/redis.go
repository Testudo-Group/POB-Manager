package database

import (
	"context"
	"crypto/tls"
	"log"
	"strings"

	"github.com/redis/go-redis/v9"
)

func ConnectRedis(url, token string) *redis.Client {
	if url == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_URL environment variable is required")
	}
	if token == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_TOKEN environment variable is required")
	}

	log.Printf("🔗 Connecting to Redis")

	// Remove protocol prefix if present
	cleanURL := strings.TrimPrefix(url, "https://")
	cleanURL = strings.TrimPrefix(cleanURL, "http://")
	
	// For Upstash, we need to add the port
	addr := cleanURL + ":6379"

	opt := &redis.Options{
		Addr:     addr,
		Password: token,
		DB:       0,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: false,
		},
	}

	client := redis.NewClient(opt)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Redis ping failed: %v", err)
		log.Printf("   Address attempted: %s", addr)
	} else {
		log.Println("✅ Redis connected successfully")
	}

	return client
}