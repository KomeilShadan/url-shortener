package redis

import (
	"github.com/go-redis/redis/v8"
	"janus/internal/config"
	"janus/pkg/log"
	"time"
)

func InitConnection(cfg *config.Config) *redis.Client {
	// Use Addr if provided, otherwise fall back to Host
	redisAddr := cfg.Redis.Addr
	if redisAddr == "" {
		redisAddr = cfg.Redis.Host
	}

	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	log.Info(log.Redis, log.Startup, "Initializing Redis connection", map[string]interface{}{
		"addr":           redisAddr,
		"db":             cfg.Redis.DB,
		"min_idle_conns": cfg.Redis.MinIdleConns,
		"pool_size":      cfg.Redis.PoolSize,
		"pool_timeout":   cfg.Redis.PoolTimeout,
	})

	client := redis.NewClient(&redis.Options{
		Addr:         redisAddr,
		MinIdleConns: cfg.Redis.MinIdleConns,
		PoolSize:     cfg.Redis.PoolSize,
		PoolTimeout:  time.Duration(cfg.Redis.PoolTimeout) * time.Second,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
	})

	return client
}
