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
	RedisURI              string
	JWTSecret             string
	AccessTokenTTLMinutes int
	RefreshTokenTTLHours  int
}

func Load() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, reading from environment")
	}

	mongoURI := getEnv("MONGO_URI", "")
	if mongoURI == "" {
		mongoURI = getEnv("MONGO_URI", "")
	}

	return &Config{
		Port:                  getEnv("PORT", "8080"),
		MongoURI:              mongoURI,
		MongoDBName:           getEnv("MONGO_DB_NAME", "pob_management"),
		RedisURI:              getEnv("REDIS_URI", "redis://localhost:6379"),
		JWTSecret:             getEnv("JWT_SECRET", "changeme"),
		AccessTokenTTLMinutes: getEnvAsInt("ACCESS_TOKEN_TTL_MINUTES", 15),
		RefreshTokenTTLHours:  getEnvAsInt("REFRESH_TOKEN_TTL_HOURS", 168),
	}
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
