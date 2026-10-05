package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Store struct {
	c *redis.Client
}

func NewStore(c *redis.Client) *Store {
	return &Store{c: c}
}

// --- refresh tokens ---

func (s *Store) SetRefreshToken(ctx context.Context, tokenID, userID string, ttl time.Duration) error {
	return s.c.Set(ctx, refreshKey(tokenID), userID, ttl).Err()
}

func (s *Store) GetRefreshToken(ctx context.Context, tokenID string) (userID string, err error) {
	val, err := s.c.Get(ctx, refreshKey(tokenID)).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("token not found")
	}
	return val, err
}

func (s *Store) DeleteRefreshToken(ctx context.Context, tokenID string) error {
	return s.c.Del(ctx, refreshKey(tokenID)).Err()
}

func (s *Store) DeleteAllUserTokens(ctx context.Context, userID string) error {
	return s.c.Del(ctx, userSessionsKey(userID)).Err()
}

// --- rate limiting (fixed window) ---

// RateLimit returns (allowed, remaining, err).
// key: e.g. "ratelimit:login:127.0.0.1"
func (s *Store) RateLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, int, error) {
	pipe := s.c.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)
	if _, err := pipe.Exec(ctx); err != nil {
		return true, limit, err
	}
	count := int(incr.Val())
	if count > limit {
		return false, 0, nil
	}
	return true, limit - count, nil
}

// --- generic helpers ---

func (s *Store) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return s.c.Set(ctx, key, value, ttl).Err()
}

func (s *Store) Get(ctx context.Context, key string) (string, error) {
	val, err := s.c.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found")
	}
	return val, err
}

func (s *Store) Del(ctx context.Context, key string) error {
	return s.c.Del(ctx, key).Err()
}

func refreshKey(tokenID string) string {
	return "refresh:" + tokenID
}

func userSessionsKey(userID string) string {
	return "sessions:" + userID
}
