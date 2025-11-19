package routes

import (
	"context"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	AppHttp "janus/internal/api/http"
	"janus/internal/config"
	"janus/internal/repository"
	"janus/internal/service"
	"janus/pkg/log"
	"net/http"
	"net/url"
	"time"
)

const (
	shortLinkTimeout = 3 * time.Second
)

// HealthCheckRoutes registers health check endpoints
func HealthCheckRoutes(router *gin.Engine) {
	router.GET("s/ping", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"message": "pong",
			"time":    time.Now().Unix(),
		})
	})

	// Detailed health check with dependencies
	router.GET("s/health", func(ctx *gin.Context) {
		ctx.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "janus",
			"version": "2.0.0",
			"time":    time.Now().Format(time.RFC3339),
		})
	})
}

// ShortLinkRoute handles short link redirection
func ShortLinkRoute(router *gin.Engine, cfg *config.Config, repos *repository.Container) {
	router.GET("/:hash", func(ctx *gin.Context) {
		shortLinkHash := ctx.Param("hash")

		// Validate hash parameter
		if shortLinkHash == "" || len(shortLinkHash) > 50 {
			log.Warn(log.Validation, log.RequestResponse, "Invalid short link hash", map[string]interface{}{
				"hash":      shortLinkHash,
				"client_ip": ctx.ClientIP(),
			})
			ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
				Message: "Bad Request",
				Error:   errors.New("invalid short link"),
				Path:    ctx.FullPath(),
			})
			return
		}

		// Try to get from cache first
		cache := service.GetLinkCache()
		var originalLink string
		var cacheHit bool

		if cache != nil {
			originalLink, cacheHit = cache.Get(shortLinkHash)
			if cacheHit {
				log.Debug(log.Redis, log.Select, "Cache hit for short link redirect", map[string]interface{}{
					"short_hash": shortLinkHash,
					"client_ip":  ctx.ClientIP(),
				})
			}
		}

		// Cache miss, fetch from database
		if !cacheHit {
			dbCtx, cancel := context.WithTimeout(ctx.Request.Context(), shortLinkTimeout)
			defer cancel()

			link, err := repos.Link.FindByHash(dbCtx, shortLinkHash)

			if err != nil {
				if errors.Is(err, mongo.ErrNoDocuments) {
					log.Warn(log.Mongodb, log.Select, "Short link not found", map[string]interface{}{
						"short_hash": shortLinkHash,
						"client_ip":  ctx.ClientIP(),
						"user_agent": ctx.Request.UserAgent(),
					})
					ctx.AbortWithStatusJSON(http.StatusNotFound, AppHttp.ApiResponse{
						Message: "Not Found",
						Error:   errors.New("link does not exist"),
						Path:    ctx.FullPath(),
					})
					return
				}

				if errors.Is(err, context.DeadlineExceeded) {
					log.Error(log.Mongodb, log.Select, err, map[string]interface{}{
						"short_hash": shortLinkHash,
						"timeout":    shortLinkTimeout.String(),
					})
					ctx.Redirect(http.StatusTemporaryRedirect, cfg.App.FallbackBaseURL)
					return
				}

				log.Error(log.Mongodb, log.Select, err, map[string]interface{}{
					"short_hash": shortLinkHash,
				})
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
					Message: "Internal Server Error",
					Error:   errors.New("database error"),
					Path:    ctx.FullPath(),
				})
				return
			}

			originalLink = link.OriginalURL

			// Store in cache for future requests
			if cache != nil {
				cache.Set(shortLinkHash, originalLink)
				log.Debug(log.Redis, log.Insert, "Link cached successfully", map[string]interface{}{
					"short_hash": shortLinkHash,
				})
			}
		}

		// Validate and parse the original URL
		if originalLink == "" {
			log.Error(log.Internal, log.Select, fmt.Errorf("empty original link"), map[string]interface{}{
				"short_hash": shortLinkHash,
			})
			ctx.Redirect(http.StatusMovedPermanently, cfg.App.FallbackBaseURL)
			return
		}

		parsedLink, err := url.Parse(originalLink)
		if err != nil {
			log.Error(log.Internal, log.Select, err, map[string]interface{}{
				"short_hash":    shortLinkHash,
				"original_link": originalLink,
			})
			// Redirect to a safe fallback
			ctx.Redirect(http.StatusMovedPermanently, cfg.App.FallbackBaseURL)
			return
		}

		// Add tracking query parameter
		linkQueryParams := parsedLink.Query()
		linkQueryParams.Set("short_link", "1")
		linkQueryParams.Set("ref", "janus")
		parsedLink.RawQuery = linkQueryParams.Encode()

		// Set custom header for tracking
		ctx.Header("X-Janus-Referrer", "short-link")
		ctx.Header("Cache-Control", "public, max-age=300") // Cache for 5 minutes

		log.Info(log.General, log.RequestResponse, "Short link redirected successfully", map[string]interface{}{
			"short_hash": shortLinkHash,
			"client_ip":  ctx.ClientIP(),
			"cache_hit":  cacheHit,
		})

		ctx.Redirect(http.StatusMovedPermanently, parsedLink.String())
	})
}
