package redis

import (
	"context"
	"fmt"
	"time"

	"tribute-back/internal/config"

	"github.com/redis/go-redis/v9"
)

var client *redis.Client

// Init initializes the Redis connection
func Init() (*redis.Client, error) {
	cfg := config.GetRedisConfig()

	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("error connecting to Redis: %w", err)
	}

	return client, nil
}

// GetClient returns the Redis client instance
func GetClient() *redis.Client {
	return client
}

// Close closes the Redis connection
func Close() error {
	if client != nil {
		return client.Close()
	}
	return nil
}

// Set sets a key-value pair in Redis
func Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return client.Set(ctx, key, value, expiration).Err()
}

// Get gets a value from Redis by key
func Get(ctx context.Context, key string) (string, error) {
	return client.Get(ctx, key).Result()
}

// Del deletes a key from Redis
func Del(ctx context.Context, key string) error {
	return client.Del(ctx, key).Err()
}

// Exists checks if a key exists in Redis
func Exists(ctx context.Context, key string) (bool, error) {
	result, err := client.Exists(ctx, key).Result()
	return result > 0, err
}
