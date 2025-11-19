package middleware

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	AppHttp "janus/internal/api/http"
	"janus/internal/config"
	"janus/pkg/log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	rateLimitDuration  = 15 * time.Minute
	rateLimitKeyPrefix = "rate_limit:"
)

var (
	mutex sync.Mutex
)

// Throttle implements rate limiting using Redis
func Throttle(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rdb == nil {
			log.Error(log.Redis, log.RequestResponse, fmt.Errorf("redis client is nil"), nil)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
				Message: "Internal Server Error",
				Error:   errors.New("rate limiting unavailable"),
				Path:    ctx.FullPath(),
			})
			return
		}

		cfg := config.Get()
		apiQuota := cfg.App.APIQuota
		clientIP := ctx.ClientIP()
		rateLimitKey := rateLimitKeyPrefix + clientIP

		mutex.Lock()
		defer mutex.Unlock()

		// Check current rate limit status
		val, err := rdb.Get(ctx.Request.Context(), rateLimitKey).Result()

		if errors.Is(err, redis.Nil) {
			// First request from this IP, initialize counter
			err = rdb.Set(ctx.Request.Context(), rateLimitKey, apiQuota, rateLimitDuration).Err()
			if err != nil {
				log.Error(log.Redis, log.Insert, err, map[string]interface{}{
					"client_ip": clientIP,
					"key":       rateLimitKey,
				})
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
					Message: "Internal Server Error",
					Error:   errors.New("rate limiting error"),
					Path:    ctx.FullPath(),
				})
				return
			}

			// Decrement counter for this request
			err = rdb.Decr(ctx.Request.Context(), rateLimitKey).Err()
			if err != nil {
				log.Error(log.Redis, log.Update, err, map[string]interface{}{
					"client_ip": clientIP,
					"key":       rateLimitKey,
				})
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
					Message: "Internal Server Error",
					Error:   errors.New("rate limiting error"),
					Path:    ctx.FullPath(),
				})
				return
			}

			log.Debug(log.Redis, log.RequestResponse, "Rate limit initialized for client", map[string]interface{}{
				"client_ip": clientIP,
				"quota":     apiQuota,
			})

			ctx.Next()
			return
		}

		if err != nil {
			log.Error(log.Redis, log.Select, err, map[string]interface{}{
				"client_ip": clientIP,
				"key":       rateLimitKey,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
				Message: "Internal Server Error",
				Error:   errors.New("rate limiting error"),
				Path:    ctx.FullPath(),
			})
			return
		}

		// Parse remaining requests
		remaining, err := strconv.Atoi(val)
		if err != nil {
			log.Error(log.Redis, log.Select, err, map[string]interface{}{
				"client_ip": clientIP,
				"value":     val,
			})
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
				Message: "Internal Server Error",
				Error:   errors.New("rate limiting error"),
				Path:    ctx.FullPath(),
			})
			return
		}

		// Check if rate limit exceeded
		if remaining <= 0 {
			ttl, err := rdb.TTL(ctx.Request.Context(), rateLimitKey).Result()
			if err != nil {
				log.Error(log.Redis, log.Select, err, map[string]interface{}{
					"client_ip": clientIP,
					"key":       rateLimitKey,
				})
				ttl = rateLimitDuration // Fallback to default
			}

			resetMinutes := int64(ttl.Minutes())
			if resetMinutes < 0 {
				resetMinutes = 0
			}

			log.Warn(log.Validation, log.RequestResponse, "Rate limit exceeded", map[string]interface{}{
				"client_ip":    clientIP,
				"path":         ctx.FullPath(),
				"reset_in_min": resetMinutes,
			})

			ctx.AbortWithStatusJSON(http.StatusTooManyRequests, AppHttp.ApiResponse{
				Message: "Too Many Requests",
				Error:   errors.New("rate limit exceeded"),
				Data: map[string]interface{}{
					"rate_limit_reset_minutes": resetMinutes,
					"retry_after":              ttl.String(),
				},
				Path: ctx.FullPath(),
			})
			return
		}

		// Decrement counter
		err = rdb.Decr(ctx.Request.Context(), rateLimitKey).Err()
		if err != nil {
			log.Error(log.Redis, log.Update, err, map[string]interface{}{
				"client_ip": clientIP,
				"key":       rateLimitKey,
			})
			// Don't block the request on decrement failure
		}

		// Add rate limit headers
		ctx.Header("X-RateLimit-Limit", apiQuota)
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(remaining-1))

		ctx.Next()
	}
}
