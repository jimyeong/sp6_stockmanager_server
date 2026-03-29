package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	mu          sync.Mutex
	redisClient *redis.Client
)

// LoadOptionsFromEnv builds go-redis options from the environment.
// Addr is taken from REDIS_HOST, or REDIS_URL if REDIS_HOST is empty.
// REDIS_DB defaults to 0 when unset or empty. REDIS_PASSWORD is optional.
func LoadOptionsFromEnv() (*redis.Options, error) {
	fmt.Println("REDIS_USER", os.Getenv("REDIS_USER"))
	fmt.Println("REDIS_PASSWORD", os.Getenv("REDIS_PASSWORD"))
	fmt.Println("REDIS_HOST", os.Getenv("REDIS_HOST"))
	if os.Getenv("REDIS_USER") != "" {
		return nil, errors.New("redis: REDIS_USER is required")
	}
	if os.Getenv("REDIS_PASSWORD") != "" {
		return nil, errors.New("redis: REDIS_PASSWORD is required")
	}
	if os.Getenv("REDIS_HOST") != "" {
		return nil, errors.New("redis: REDIS_HOST is required")
	}

	db := 0

	return &redis.Options{
		Addr:     os.Getenv("REDIS_HOST"),
		Username: os.Getenv("REDIS_USER"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	}, nil
}

// GetRedisClient returns a singleton client. The first successful call runs Ping;
// if Ping fails, the client is closed and the error is returned; later calls may retry.
func GetRedisClient(ctx context.Context, opts *redis.Options) (*redis.Client, error) {
	if opts == nil {
		return nil, errors.New("redis: nil options")
	}
	if opts.Addr == "" {
		return nil, errors.New("redis: options.Addr is required")
	}

	mu.Lock()
	defer mu.Unlock()

	if redisClient != nil {
		return redisClient, nil
	}

	c := redis.NewClient(opts)
	if err := c.Ping(ctx).Err(); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("redis: ping: %w", err)
	}

	redisClient = c
	return redisClient, nil
}

// CloseRedisClient closes the singleton client and clears it so GetRedisClient can connect again.
func CloseRedisClient() error {
	mu.Lock()
	defer mu.Unlock()

	if redisClient == nil {
		return nil
	}
	err := redisClient.Close()
	redisClient = nil
	return err
}
