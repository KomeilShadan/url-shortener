package rest

import (
	"drto-link/internal/api/routes"
	"drto-link/internal/config"
	"drto-link/pkg/log"
	"github.com/gin-gonic/gin"
	"strconv"
)

var router *gin.Engine

func InitServer(cfg *config.Config) {
	mode := cfg.App.Mode
	if mode != gin.DebugMode && mode != gin.TestMode {
		mode = gin.ReleaseMode
	}

	//Default returns an Engine instance with the Logger and Recovery middleware already attached.
	//Recovery returns a middleware that recovers from any panics and writes a 500 if there was one.
	router = gin.Default()

	routes.HealthCheckRoutes(router)
	routes.ApiRoutes(router)

	err := router.Run(":" + strconv.Itoa(cfg.App.Port))

	if err != nil {
		log.Error(log.General, log.Startup, err, nil)
		panic(err)
	}
}
