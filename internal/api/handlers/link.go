package handlers

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
	AppHttp "janus/internal/api/http"
	"janus/internal/api/request"
	"janus/internal/api/response"
	"janus/internal/config"
	"janus/internal/repository"
	"janus/internal/service"
	"janus/internal/utils"
	"janus/pkg/log"
	"net/http"
	"strings"
	"time"
)

const (
	requestTimeout = 5 * time.Second
)

// LinkHandler handles link-related requests
type LinkHandler struct {
	linkRepo repository.LinkRepository
}

// NewLinkHandler creates a new link handler
func NewLinkHandler(linkRepo repository.LinkRepository) *LinkHandler {
	if linkRepo == nil {
		panic("linkRepo cannot be nil")
	}
	return &LinkHandler{linkRepo: linkRepo}
}

func (linkHandler *LinkHandler) ShortLink(ctx *gin.Context) {
	cfg := config.Get()
	repo := linkHandler.linkRepo

	var (
		req              request.ShortLinkRequest
		link             string
		shortLinkBaseURL string
		fullShortLink    string
	)

	// Bind the JSON payload to the ShortLinkRequest struct
	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn(log.Validation, log.RequestResponse, "Invalid request body", map[string]interface{}{
			"error":     err.Error(),
			"client_ip": ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
			Message: "Bad Request",
			Error:   errors.New("malformed request body"),
			Path:    ctx.FullPath(),
		})
		return
	}

	// Validate the ShortLinkRequest struct
	utils.ValidateRequestBody(ctx, &req)
	if ctx.IsAborted() {
		return
	}

	log.Info(log.General, log.RequestResponse, "Received short link creation request", map[string]interface{}{
		"path":       ctx.FullPath(),
		"client_ip":  ctx.ClientIP(),
		"link":       req.Link,
		"expirable":  req.Expirable,
		"user_agent": ctx.Request.UserAgent(),
	})

	// Validate domain to prevent self-referencing
	if !utils.AvoidDSelfDomain(req.Link) {
		log.Warn(log.Validation, log.RequestResponse, "Self-domain link rejected", map[string]interface{}{
			"link":      req.Link,
			"client_ip": ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, AppHttp.ApiResponse{
			Message: "Unprocessable Entity (Nice Try!)",
			Error:   errors.New("unprocessable input link"),
			Path:    ctx.FullPath(),
		})
		return
	}

	link = utils.EnforceHTTP(req.Link)
	shortLinkBaseURL = cfg.App.ShortLinkBaseURL

	// Generate short link hash
	shortLinkHash, err := service.GenerateShortLinkHash(link)
	if err != nil {
		log.Error(log.Internal, log.Insert, err, map[string]interface{}{
			"link": link,
		})
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("failed to generate short link"),
			Path:    ctx.FullPath(),
		})
		return
	}

	fullShortLink = shortLinkBaseURL + shortLinkHash

	// Create context with timeout for database operation
	dbCtx, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()

	// Insert link into database synchronously to ensure data consistency
	err = repo.Insert(dbCtx, link, shortLinkHash, time.Now())
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error(log.Mongodb, log.Insert, err, map[string]interface{}{
				"link":       link,
				"short_hash": shortLinkHash,
				"timeout":    requestTimeout.String(),
			})
			ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, AppHttp.ApiResponse{
				Message: "Request Timeout",
				Error:   errors.New("database operation timed out"),
				Path:    ctx.FullPath(),
			})
			return
		}

		log.Error(log.Mongodb, log.Insert, err, map[string]interface{}{
			"link":       link,
			"short_hash": shortLinkHash,
		})
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
		return
	}

	log.Info(log.Mongodb, log.Insert, "Short link created successfully", map[string]interface{}{
		"short_hash": shortLinkHash,
		"link":       link,
		"client_ip":  ctx.ClientIP(),
	})

	ctx.JSON(http.StatusCreated, AppHttp.ApiResponse{
		Data: response.ShortLinkResponse{
			Link:      link,
			ShortLink: fullShortLink,
			Expirable: req.Expirable,
		},
	})
}

func (linkHandler *LinkHandler) ResolveLink(ctx *gin.Context) {
	cfg := config.Get()
	repo := linkHandler.linkRepo

	shortLink := ctx.Request.URL.Path
	shortLinkHash := strings.TrimPrefix(shortLink, cfg.App.ShortLinkBaseURL)

	if shortLinkHash == "" {
		log.Warn(log.Validation, log.RequestResponse, "Empty short link hash", map[string]interface{}{
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
	if cache != nil {
		if originalLink, found := cache.Get(shortLinkHash); found {
			log.Debug(log.Redis, log.Select, "Cache hit for short link", map[string]interface{}{
				"short_hash": shortLinkHash,
			})
			ctx.Redirect(http.StatusMovedPermanently, originalLink)
			return
		}
	}

	// Create context with timeout for database operation
	dbCtx, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()

	// Cache miss, fetch from database
	link, err := repo.FindByHash(dbCtx, shortLinkHash)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn(log.Mongodb, log.Select, "Short link not found", map[string]interface{}{
				"short_hash": shortLinkHash,
				"client_ip":  ctx.ClientIP(),
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
				"timeout":    requestTimeout.String(),
			})
			ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, AppHttp.ApiResponse{
				Message: "Request Timeout",
				Error:   errors.New("database operation timed out"),
				Path:    ctx.FullPath(),
			})
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

	// Store in cache for future requests
	if cache != nil {
		cache.Set(shortLinkHash, link.OriginalURL)
	}

	log.Info(log.Mongodb, log.Select, "Short link resolved successfully", map[string]interface{}{
		"short_hash": shortLinkHash,
		"client_ip":  ctx.ClientIP(),
	})

	ctx.Redirect(http.StatusMovedPermanently, link.OriginalURL)
}

func (linkHandler *LinkHandler) GetLink(ctx *gin.Context) {
	cfg := config.Get()
	repo := linkHandler.linkRepo

	var req request.GetLinkRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn(log.Validation, log.RequestResponse, "Invalid request body", map[string]interface{}{
			"error":     err.Error(),
			"client_ip": ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
			Message: "Bad Request",
			Error:   errors.New("malformed request body"),
			Path:    ctx.FullPath(),
		})
		return
	}

	utils.ValidateRequestBody(ctx, &req)
	if ctx.IsAborted() {
		return
	}

	shortLinkHash := strings.TrimPrefix(req.ShortLink, cfg.App.ShortLinkBaseURL)

	if shortLinkHash == "" {
		log.Warn(log.Validation, log.RequestResponse, "Empty short link hash", map[string]interface{}{
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
	if cache != nil {
		if originalLink, found := cache.Get(shortLinkHash); found {
			log.Debug(log.Redis, log.Select, "Cache hit for link retrieval", map[string]interface{}{
				"short_hash": shortLinkHash,
			})
			ctx.JSON(http.StatusOK, AppHttp.ApiResponse{
				Data: response.ResolveLinkResponse{
					Link: originalLink,
				},
			})
			return
		}
	}

	// Create context with timeout for database operation
	dbCtx, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()

	// Cache miss, fetch from database
	link, err := repo.FindByHash(dbCtx, shortLinkHash)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			log.Warn(log.Mongodb, log.Select, "Link not found", map[string]interface{}{
				"short_hash": shortLinkHash,
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
				"timeout":    requestTimeout.String(),
			})
			ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, AppHttp.ApiResponse{
				Message: "Request Timeout",
				Error:   errors.New("database operation timed out"),
				Path:    ctx.FullPath(),
			})
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

	// Store in cache for future requests
	if cache != nil {
		cache.Set(shortLinkHash, link.OriginalURL)
	}

	ctx.JSON(http.StatusOK, AppHttp.ApiResponse{
		Data: response.ResolveLinkResponse{
			Link: link.OriginalURL,
		},
	})
}

func (linkHandler *LinkHandler) UpdateLink(ctx *gin.Context) {
	repo := linkHandler.linkRepo

	var (
		req  request.UpdateLinkRequest
		link string
	)

	if err := ctx.ShouldBindJSON(&req); err != nil {
		log.Warn(log.Validation, log.RequestResponse, "Invalid request body", map[string]interface{}{
			"error":     err.Error(),
			"client_ip": ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusBadRequest, AppHttp.ApiResponse{
			Message: "Bad Request",
			Error:   errors.New("malformed request body"),
			Path:    ctx.FullPath(),
		})
		return
	}

	utils.ValidateRequestBody(ctx, &req)
	if ctx.IsAborted() {
		return
	}

	if !utils.AvoidDSelfDomain(req.Link) {
		log.Warn(log.Validation, log.RequestResponse, "Self-domain link rejected in update", map[string]interface{}{
			"link":       req.Link,
			"short_link": req.ShortLink,
			"client_ip":  ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusUnprocessableEntity, AppHttp.ApiResponse{
			Message: "Unprocessable Entity (Nice Try!)",
			Error:   errors.New("unprocessable input link"),
			Path:    ctx.FullPath(),
		})
		return
	}

	link = utils.EnforceHTTP(req.Link)

	// Create context with timeout for database operation
	dbCtx, cancel := context.WithTimeout(ctx.Request.Context(), requestTimeout)
	defer cancel()

	matchedCount, err := repo.Update(dbCtx, req.ShortLink, link)

	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			log.Error(log.Mongodb, log.Update, err, map[string]interface{}{
				"short_link": req.ShortLink,
				"new_link":   link,
				"timeout":    requestTimeout.String(),
			})
			ctx.AbortWithStatusJSON(http.StatusGatewayTimeout, AppHttp.ApiResponse{
				Message: "Request Timeout",
				Error:   errors.New("database operation timed out"),
				Path:    ctx.FullPath(),
			})
			return
		}

		log.Error(log.Mongodb, log.Update, err, map[string]interface{}{
			"short_link": req.ShortLink,
			"new_link":   link,
		})
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, AppHttp.ApiResponse{
			Message: "Internal Server Error",
			Error:   errors.New("database error"),
			Path:    ctx.FullPath(),
		})
		return
	}

	if matchedCount == 0 {
		log.Warn(log.Mongodb, log.Update, "Link not found for update", map[string]interface{}{
			"short_link": req.ShortLink,
			"client_ip":  ctx.ClientIP(),
		})
		ctx.AbortWithStatusJSON(http.StatusNotFound, AppHttp.ApiResponse{
			Message: "Not Found",
			Error:   errors.New("link does not exist"),
			Path:    ctx.FullPath(),
		})
		return
	}

	// Invalidate cache since link was updated
	cache := service.GetLinkCache()
	if cache != nil {
		cache.Delete(req.ShortLink)
		log.Debug(log.Redis, log.Update, "Cache invalidated for updated link", map[string]interface{}{
			"short_link": req.ShortLink,
		})
	}

	log.Info(log.Mongodb, log.Update, "Link updated successfully", map[string]interface{}{
		"short_link": req.ShortLink,
		"new_link":   link,
		"client_ip":  ctx.ClientIP(),
	})

	ctx.JSON(http.StatusAccepted, AppHttp.ApiResponse{
		Data: response.UpdateLinkResponse{
			Link:      link,
			ShortLink: req.ShortLink,
		},
	})
}
