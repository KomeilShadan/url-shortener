package repository

import (
	"context"
	"janus/internal/model"
	"time"
)

// LinkRepository defines the interface for link data operations
type LinkRepository interface {
	Insert(ctx context.Context, link, shortLinkHash string, timestamp time.Time) error
	FindByHash(ctx context.Context, shortLinkHash string) (*model.Link, error)
	Update(ctx context.Context, shortLinkHash, newLink string) (int64, error)
	Delete(ctx context.Context, shortLinkHash string) error
}

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
}
