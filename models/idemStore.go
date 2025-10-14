package models

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	headerIdempotencyKey = "Idempotency-Key"
	defaultTTL           = 24 * time.Hour
)

type IdemStore struct {
	Rdb            *redis.Client
	TTL            time.Duration
	acquireScript  *redis.Script
	completeScript *redis.Script
}

func NewIdemStore(rdb *redis.Client, ttl time.Duration, acquireLua string, completeLua string) *IdemStore {
	return &IdemStore{
		Rdb:            rdb,
		TTL:            ttl,
		acquireScript:  redis.NewScript(acquireLua),
		completeScript: redis.NewScript(completeLua),
	}
}

func (s *IdemStore) Acquire(ctx context.Context, key string, reqHash string) (state string, status string, body string, error error) {
	res, err := s.acquireScript.Run(ctx, s.Rdb, []string{"idem: " + key}, int(s.TTL/time.Second), reqHash).Result()
	if err != nil {
		return "", "", "", err
	}
	arr := res.([]interface{})
	return arr[0].(string), arr[1].(string), arr[2].(string), nil
}

func (s *IdemStore) Complete(ctx context.Context, key string, status int, body string) error {
	_, err := s.completeScript.Run(ctx, s.Rdb, []string{"idem: " + key}, strconv.Itoa(status), body, int(s.TTL/time.Second)).Result()
	return err
}
