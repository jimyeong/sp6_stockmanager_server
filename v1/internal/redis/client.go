package redis

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/redis/go-redis/v9"
)

var (
	mu          sync.Mutex
	redisClient *redis.Client
)

// LoadOptionsFromEnv builds go-redis options from the environment.
// Addr is taken from REDIS_ADDR, or REDIS_URL if REDIS_ADDR is empty.
// REDIS_DB defaults to 0 when unset or empty. REDIS_PASSWORD is optional.
func LoadOptionsFromEnv() (*redis.Options, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = os.Getenv("REDIS_URL")
	}
	if addr == "" {
		return nil, errors.New("redis: set REDIS_ADDR or REDIS_URL")
	}

	db := 0
	if raw := os.Getenv("REDIS_DB"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			return nil, fmt.Errorf("redis: invalid REDIS_DB: %w", err)
		}
		db = parsed
	}

	return &redis.Options{
		Addr:     addr,
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
