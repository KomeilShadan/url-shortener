package routes

import (
	"drto-link/internal/api/handlers"
	"drto-link/internal/api/middleware"
	"drto-link/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
)

var api *gin.RouterGroup

func ApiRoutes(router *gin.Engine, cfg *config.Config, mongo *mongo.Client, rdb *redis.Client) {
	api = router.Group("/api")

	api.Use(middleware.AuthMiddleware()).
		//Use(middleware.Throttle(rdb)).
		Use(middleware.InjectMongoClient(mongo)).
		Use(middleware.InjectRedisClient(rdb)).
		POST("link/short", handlers.ShortLink).
		GET("link/resolve", handlers.ResolveLink).
		PUT("link/", handlers.UpdateLink)
}
