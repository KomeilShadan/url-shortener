package rest

import (
	"drto-link/internal/api/routes"
	"drto-link/internal/config"
	"drto-link/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"strconv"
)

var (
	router *gin.Engine
)

func InitServer(cfg *config.Config, mongo *mongo.Client, rdb *redis.Client) {
	mode := cfg.App.Mode
	if mode != gin.DebugMode && mode != gin.TestMode {
		mode = gin.ReleaseMode
	}
	gin.SetMode(mode)

	//Default returns an Engine instance with the Logger and Recovery middleware already attached.
	//Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
	router = gin.Default()

	routes.HealthCheckRoutes(router, cfg)
	routes.ApiRoutes(router, cfg, mongo, rdb)

	err := router.Run(":" + strconv.Itoa(cfg.App.Port))

	if err != nil {
		log.Error(log.General, log.Startup, err, nil)
		panic(err)
	}
}
