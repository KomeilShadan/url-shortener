package routes

import (
	"drto-link/internal/api/middleware"
	"drto-link/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

var api *gin.RouterGroup

func ApiRoutes(router *gin.Engine, cfg *config.Config, rdb *redis.Client, mongo *mongo.Client) {
	api = router.Group("/api")
	api.Use(middleware.AuthMiddleware())
	api.Use(middleware.Throttle(rdb))
	//api.GET("/", func(ctx *gin.Context) {})
}
