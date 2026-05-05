package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                  string
	MongoURI              string
	MongoDBName           string
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int

	// 🔥 Upstash REST Redis (NOT TCP)
	UpstashRedisURL   string
	UpstashRedisToken string
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment")
	}

	config := &Config{
		Port:        getEnv("PORT", "10000"),
		MongoURI:    getEnv("MONGO_URI", ""),
		MongoDBName: getEnv("MONGO_DB_NAME", "pob_management"),

		JWTSecret:             getEnv("JWT_SECRET", "changeme"),
		AccessTokenTTLMinutes: getEnvAsInt("ACCESS_TOKEN_TTL_MINUTES", 15),
		RefreshTokenTTLHours:  getEnvAsInt("REFRESH_TOKEN_TTL_HOURS", 168),

		// ✅ Upstash REST credentials
		UpstashRedisURL:   getEnv("UPSTASH_REDIS_REST_URL", ""),
		UpstashRedisToken: getEnv("UPSTASH_REDIS_REST_TOKEN", ""),
	}

	// Validate required configuration
	config.validate()
	
	return config
}

// validate checks if all required configuration fields are set
func (c *Config) validate() {
	// Check MongoDB
	if c.MongoURI == "" {
		log.Fatal("❌ MONGO_URI environment variable is required")
	}
	
	// Check Redis (Upstash)
	if c.UpstashRedisURL == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_URL environment variable is required\n   Get one from https://console.upstash.com/")
	}
	
	if c.UpstashRedisToken == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_TOKEN environment variable is required\n   Get one from https://console.upstash.com/")
	}
	
	// Check JWT Secret in production
	environment := getEnv("ENVIRONMENT", "development")
	if c.JWTSecret == "changeme" && environment == "production" {
		log.Fatal("❌ JWT_SECRET must be changed from default in production")
	}
	
	// Log success (without exposing secrets)
	log.Printf("✅ Configuration loaded successfully")
	log.Printf("   - Port: %s", c.Port)
	log.Printf("   - MongoDB Database: %s", c.MongoDBName)
	log.Printf("   - Redis URL: %s", c.UpstashRedisURL)
	log.Printf("   - Environment: %s", environment)
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getEnvAsInt(key string, fallback int) int {
	value, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return parsed
}
