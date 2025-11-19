package repository

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
	"time"
)

// redisCacheRepository implements CacheRepository interface
type redisCacheRepository struct {
	client *redis.Client
}

// NewRedisCacheRepository creates a new Redis cache repository instance
func NewRedisCacheRepository(client *redis.Client) CacheRepository {
	return &redisCacheRepository{
		client: client,
	}
}

// Get retrieves a value from cache
func (r *redisCacheRepository) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", nil // Key doesn't exist
	}
	return val, err
}

// Set stores a value in cache with TTL
func (r *redisCacheRepository) Set(ctx context.Context, key, value string, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a value from cache
func (r *redisCacheRepository) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *redisCacheRepository) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}
