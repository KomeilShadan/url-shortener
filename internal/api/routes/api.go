package routes

import (
	"github.com/gin-gonic/gin"
	"janus/internal/api/handlers"
	"janus/internal/api/middleware"
	"janus/internal/config"
	"janus/internal/repository"
)

var api *gin.RouterGroup

func ApiRoutes(router *gin.Engine, cfg *config.Config, repos *repository.Container) {
	api = router.Group("/api/links")

	// Initialize handlers with repositories
	linkHandler := handlers.NewLinkHandler(repos.Link)

	api.Use(middleware.AuthMiddleware()).
		//Use(middleware.Throttle(rdb)).
		POST("/short", linkHandler.ShortLink).
		GET("/resolve", linkHandler.ResolveLink).
		GET("/", linkHandler.GetLink).
		PUT("/", linkHandler.UpdateLink)
}
