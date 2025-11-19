package rest

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"janus/internal/api/routes"
	"janus/internal/config"
	"janus/internal/repository"
	"janus/internal/service"
	"janus/pkg/log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"
)

const (
	defaultReadTimeout   = 10 * time.Second
	defaultWriteTimeout  = 10 * time.Second
	defaultIdleTimeout   = 120 * time.Second
	shutdownTimeout      = 15 * time.Second
	maxHeaderBytes       = 1 << 20 // 1 MB
	cacheTTL             = 5 * time.Minute
	cacheCleanupInterval = 1 * time.Minute
	cacheMaxSize         = 10000
)

var (
	router *gin.Engine
)

func InitServer(cfg *config.Config, mongo *mongo.Client, rdb *redis.Client) {
	// Set Gin mode based on configuration
	mode := cfg.App.Mode
	if mode != gin.DebugMode && mode != gin.TestMode {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	log.Info(log.General, log.Startup, "Initializing server components", map[string]interface{}{
		"mode": mode,
	})

	// Initialize repositories with validation
	repos := repository.InitRepositories(mongo, rdb, cfg.Mongo.DB)
	if repos == nil {
		log.Error(log.Internal, log.Startup, fmt.Errorf("failed to initialize repositories"), nil)
		panic("repository initialization failed")
	}
	log.Info(log.Internal, log.Startup, "Repositories initialized successfully", nil)

	// Initialize in-memory cache for frequently accessed links
	cache := service.InitLinkCache(service.LinkCacheConfig{
		TTL:             cacheTTL,
		CleanupInterval: cacheCleanupInterval,
		MaxSize:         cacheMaxSize,
	})
	if cache == nil {
		log.Error(log.Internal, log.Startup, fmt.Errorf("failed to initialize link cache"), nil)
		panic("cache initialization failed")
	}
	log.Info(log.Internal, log.Startup, "Link cache initialized successfully", map[string]interface{}{
		"ttl":              cacheTTL.String(),
		"cleanup_interval": cacheCleanupInterval.String(),
		"max_size":         cacheMaxSize,
	})

	// Initialize Gin router with custom recovery middleware
	router = gin.New()
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		SkipPaths: []string{"/s/ping"}, // Skip health check logs
	}))
	router.Use(customRecoveryMiddleware())

	// Register routes
	routes.HealthCheckRoutes(router)
	routes.ApiRoutes(router, cfg, repos)
	routes.ShortLinkRoute(router, cfg, repos)

	// Create HTTP server with production-grade timeouts
	srv := &http.Server{
		Addr:           ":" + strconv.Itoa(cfg.App.Port),
		Handler:        router,
		ReadTimeout:    defaultReadTimeout,
		WriteTimeout:   defaultWriteTimeout,
		IdleTimeout:    defaultIdleTimeout,
		MaxHeaderBytes: maxHeaderBytes,
	}

	// Channel to capture server errors
	serverErrors := make(chan error, 1)

	// Start server in a goroutine
	go func() {
		log.Info(log.General, log.Startup, "HTTP server starting", map[string]interface{}{
			"addr":          srv.Addr,
			"read_timeout":  defaultReadTimeout.String(),
			"write_timeout": defaultWriteTimeout.String(),
			"idle_timeout":  defaultIdleTimeout.String(),
		})

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrors <- err
		}
	}()

	log.Info(log.General, log.Startup, "Server started successfully", map[string]interface{}{
		"port": cfg.App.Port,
		"mode": mode,
		"host": cfg.App.Host,
	})

	// Wait for interrupt signal or server error
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		log.Error(log.General, log.Startup, err, map[string]interface{}{
			"port": cfg.App.Port,
		})
		panic(err)
	case sig := <-quit:
		log.Info(log.General, log.Shutdown, "Shutdown signal received", map[string]interface{}{
			"signal": sig.String(),
		})
	}

	// Graceful shutdown
	log.Info(log.General, log.Shutdown, "Initiating graceful shutdown", map[string]interface{}{
		"timeout": shutdownTimeout.String(),
	})

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error(log.General, log.Shutdown, err, map[string]interface{}{
			"timeout": shutdownTimeout.String(),
		})
	} else {
		log.Info(log.General, log.Shutdown, "Server shutdown completed successfully", nil)
	}
}

// customRecoveryMiddleware provides custom panic recovery with detailed logging
func customRecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Error(log.Internal, log.RequestResponse, fmt.Errorf("panic recovered: %v", err), map[string]interface{}{
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"client_ip":  c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				})

				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Internal Server Error",
					"error":   "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}
