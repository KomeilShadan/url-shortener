package routes

import (
	"drto-link/internal/config"
	"github.com/gin-gonic/gin"
)

func HealthCheckRoutes(router *gin.Engine, cfg *config.Config) {
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
