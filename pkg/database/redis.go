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

type UpstashRedis struct {
	URL    string
	Token  string
	Client *http.Client
}

// RedisInterface defines the methods our app uses
type RedisInterface interface {
	Set(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, key string) error
	Ping(ctx context.Context) error
}

func ConnectRedis(url, token string) *UpstashRedis {
	// Validate inputs - fail fast if missing
	if url == "" {
		log.Fatal("❌ Redis URL is required but was empty. Set UPSTASH_REDIS_REST_URL environment variable")
	}
	if token == "" {
		log.Fatal("❌ Redis Token is required but was empty. Set UPSTASH_REDIS_REST_TOKEN environment variable")
	}
	
	log.Printf("🔗 Connecting to Upstash Redis at: %s", url)
	
	return &UpstashRedis{
		URL:   url,
		Token: token,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (r *UpstashRedis) Set(ctx context.Context, key, value string) error {
	return r.exec(ctx, "SET", key, value)
}

func (r *UpstashRedis) Get(ctx context.Context, key string) (string, error) {
	var result string
	err := r.exec(ctx, "GET", key, &result)
	return result, err
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
	body := []interface{}{command}
	body = append(body, args...)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("failed to marshal Redis command: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", r.URL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return fmt.Errorf("failed to create Redis request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", r.Token))

	resp, err := r.Client.Do(req)
	if err != nil {
		return fmt.Errorf("Redis request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Redis error: status %d - check your URL and token", resp.StatusCode)
	}

	var result []interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("failed to decode Redis response: %w", err)
	}

	if len(result) > 0 && len(args) > 0 {
		if ptr, ok := args[len(args)-1].(*string); ok && result[0] != nil {
			*ptr = fmt.Sprintf("%v", result[0])
		}
	}

	return nil
}
