package routes

import (
	"drto-link/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

var api *gin.RouterGroup

func ApiRoutes(router *gin.Engine) {
	api = router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	//api.GET("/", func(ctx *gin.Context) {})
}
