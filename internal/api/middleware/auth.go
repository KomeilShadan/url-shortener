package middleware

import (
	"errors"
	"github.com/gin-gonic/gin"
	AppHttp "janus/internal/api/http"
	"janus/internal/config"
	"janus/pkg/log"
	"net/http"
	"strings"
)

const (
	apiKeyHeader = "x-api-key"
)

// AuthMiddleware validates API key for protected endpoints
func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := config.Get()

		// Skip auth if API key is not configured (development mode)
		if cfg.Link.ApiKey == "" {
			log.Warn(log.General, log.RequestResponse, "API key not configured, skipping authentication", map[string]interface{}{
				"path": ctx.FullPath(),
			})
			ctx.Next()
			return
		}

		requestApiKey := strings.TrimSpace(ctx.GetHeader(apiKeyHeader))

		if requestApiKey == "" {
			log.Warn(log.Validation, log.RequestResponse, "Missing API key", map[string]interface{}{
				"client_ip":  ctx.ClientIP(),
				"path":       ctx.FullPath(),
				"user_agent": ctx.Request.UserAgent(),
			})
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AppHttp.ApiResponse{
				Message: "Unauthorized",
				Error:   errors.New("API key is required"),
				Path:    ctx.FullPath(),
			})
			return
		}

		if requestApiKey != cfg.Link.ApiKey {
			log.Warn(log.Validation, log.RequestResponse, "Invalid API key", map[string]interface{}{
				"client_ip":  ctx.ClientIP(),
				"path":       ctx.FullPath(),
				"user_agent": ctx.Request.UserAgent(),
			})
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AppHttp.ApiResponse{
				Message: "Unauthorized",
				Error:   errors.New("invalid API key"),
				Path:    ctx.FullPath(),
			})
			return
		}

		// API key is valid, proceed
		ctx.Next()
	}
}
