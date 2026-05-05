package database

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

// RedisInterface defines the methods our app uses
type RedisInterface interface {
	Set(ctx context.Context, key string, value string) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

type UpstashRedis struct {
	URL    string
	Token  string
	Client *http.Client
}

// Ensure UpstashRedis implements RedisInterface at compile time
var _ RedisInterface = (*UpstashRedis)(nil)

func ConnectRedis(url, token string) *UpstashRedis {
	// Validate inputs
	if url == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_URL environment variable is required")
	}
	if token == "" {
		log.Fatal("❌ UPSTASH_REDIS_REST_TOKEN environment variable is required")
	}
	
	log.Printf("🔗 Connecting to Upstash Redis at: %s", url)
	
	return &UpstashRedis{
		URL:   url,
		Token: token,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (r *UpstashRedis) Set(ctx context.Context, key string, value string) error {
	return r.exec(ctx, "SET", key, value)
}

func (r *UpstashRedis) Get(ctx context.Context, key string) (string, error) {
	var result string
	err := r.exec(ctx, "GET", key, &result)
	if err != nil {
		return "", err
	}
	
	// Handle nil result from Redis
	if result == "" || result == "<nil>" {
		return "", nil
	}
	
	return result, nil
}

func (r *UpstashRedis) Del(ctx context.Context, key string) error {
	return r.exec(ctx, "DEL", key)
}

func (r *UpstashRedis) Ping(ctx context.Context) error {
	log.Println("🏓 Pinging Upstash Redis...")
	err := r.exec(ctx, "PING")
	if err == nil {
		log.Println("✅ Redis connection successful!")
	}
	return err
}

func (r *UpstashRedis) exec(ctx context.Context, command string, args ...interface{}) error {
	// Build the Redis command
	body := []interface{}{command}
	body = append(body, args...)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal Redis command: %w", err)
	}

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", r.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create Redis request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.Token))

	// Execute request
	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Redis request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Redis error: status %d - check your URL and token", resp.StatusCode)
	}

	// Parse response
	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode Redis response: %w", err)
	}

	// Handle string result for GET command
	if len(result) > 0 && len(args) > 0 {
		if ptr, ok := args[len(args)-1].(*string); ok && result[0] != nil {
			*ptr = fmt.Sprintf("%v", result[0])
		}
	}

	return nil
}
