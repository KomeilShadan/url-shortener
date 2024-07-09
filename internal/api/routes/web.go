package routes

import "github.com/gin-gonic/gin"

func HealthCheckRoutes(router *gin.Engine) {
	router.GET("/ping", func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
