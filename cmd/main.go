package main

import (
	"context"
	"fmt"
	"github.com/caarlos0/env/v11"
	"github.com/getsentry/sentry-go"
	api "janus/internal/api"
	"janus/internal/config"
	"janus/pkg/log"
	"janus/pkg/mongodb"
	"janus/pkg/redis"
	stdLog "log"
	"os"
	"runtime/debug"
	"time"
)

const (
	initTimeout     = 15 * time.Second
	sentryFlushTime = 2 * time.Second
)

func main() {
	// Global panic recovery with stack trace
	defer func() {
		if r := recover(); r != nil {
			stdLog.Printf("FATAL: Application panic recovered: %v\nStack trace:\n%s", r, debug.Stack())
			sentry.CurrentHub().Recover(r)
			sentry.Flush(sentryFlushTime)
			os.Exit(1)
		}
	}()

	// Load and parse configuration
	cfg := config.Get()
	if err := env.Parse(cfg); err != nil {
		stdLog.Fatalf("FATAL: Failed to parse configuration: %v", err)
	}

	// Validate critical configuration
	if err := validateConfig(cfg); err != nil {
		stdLog.Fatalf("FATAL: Invalid configuration: %v", err)
	}

	// Initialize logger first
	_ = log.GetLogger() // Initialize the singleton logger
	log.Info(log.General, log.Startup, "Starting Janus service", map[string]interface{}{
		"environment": cfg.App.Environment,
		"mode":        cfg.App.Mode,
		"port":        cfg.App.Port,
	})

	// Initialize Sentry for error tracking
	if cfg.Sentry.Dsn != "" {
		if err := sentry.Init(sentry.ClientOptions{
			Dsn:              cfg.Sentry.Dsn,
			Environment:      cfg.Sentry.Environment,
			TracesSampleRate: 1.0,
			AttachStacktrace: true,
		}); err != nil {
			log.Error(log.Sentry, log.Startup, err, map[string]interface{}{
				"dsn": cfg.Sentry.Dsn,
			})
			// Don't exit, continue without Sentry
		} else {
			log.Info(log.Sentry, log.Startup, "Sentry initialized successfully", nil)
		}
		defer sentry.Flush(sentryFlushTime)
	}

	// Create a context with timeout for initialization
	ctx, cancel := context.WithTimeout(context.Background(), initTimeout)
	defer cancel()

	// Initialize MongoDB connection with retry
	mongo, err := initMongoWithRetry(ctx, cfg, 3)
	if err != nil {
		log.Error(log.Mongodb, log.Startup, err, map[string]interface{}{
			"uri": cfg.Mongo.URI,
		})
		os.Exit(1)
	}
	defer func() {
		disconnectCtx, disconnectCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer disconnectCancel()
		if err := mongo.Disconnect(disconnectCtx); err != nil {
			log.Error(log.Mongodb, log.Shutdown, err, nil)
		} else {
			log.Info(log.Mongodb, log.Shutdown, "MongoDB disconnected successfully", nil)
		}
	}()

	// Initialize Redis connection
	rdb := redis.InitConnection(cfg)
	if rdb == nil {
		log.Error(log.Redis, log.Startup, fmt.Errorf("failed to initialize Redis client"), map[string]interface{}{
			"addr": cfg.Redis.Addr,
		})
		os.Exit(1)
	}

	// Test Redis connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Error(log.Redis, log.Startup, err, map[string]interface{}{
			"addr": cfg.Redis.Addr,
		})
		os.Exit(1)
	}
	log.Info(log.Redis, log.Startup, "Redis connected successfully", map[string]interface{}{
		"addr": cfg.Redis.Addr,
	})

	defer func() {
		if err := rdb.Close(); err != nil {
			log.Error(log.Redis, log.Shutdown, err, nil)
		} else {
			log.Info(log.Redis, log.Shutdown, "Redis disconnected successfully", nil)
		}
	}()

	// Start the REST server (blocks until shutdown)
	api.InitServer(cfg, mongo, rdb)

	log.Info(log.General, log.Shutdown, "Application shutdown complete", nil)
}

// validateConfig validates critical configuration values
func validateConfig(cfg *config.Config) error {
	if cfg.App.Port <= 0 || cfg.App.Port > 65535 {
		return fmt.Errorf("invalid port number: %d", cfg.App.Port)
	}
	if cfg.Mongo.URI == "" {
		return fmt.Errorf("MongoDB URI is required")
	}
	if cfg.Mongo.DB == "" {
		return fmt.Errorf("MongoDB database name is required")
	}
	if cfg.Redis.Addr == "" {
		return fmt.Errorf("Redis address is required")
	}
	if cfg.App.ShortLinkBaseURL == "" {
		return fmt.Errorf("short link base URL is required")
	}
	return nil
}

// initMongoWithRetry attempts to connect to MongoDB with retries
func initMongoWithRetry(ctx context.Context, cfg *config.Config, maxRetries int) (*mongodb.Client, error) {
	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			log.Info(log.Mongodb, log.Startup, fmt.Sprintf("Retrying MongoDB connection (attempt %d/%d)", i+1, maxRetries), nil)
			time.Sleep(time.Duration(i) * 2 * time.Second)
		}

		mongo, err := mongodb.InitConnection(ctx, cfg)
		if err == nil {
			log.Info(log.Mongodb, log.Startup, "MongoDB connected successfully", map[string]interface{}{
				"database": cfg.Mongo.DB,
			})
			return mongo, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("failed to connect to MongoDB after %d attempts: %w", maxRetries, lastErr)
}
