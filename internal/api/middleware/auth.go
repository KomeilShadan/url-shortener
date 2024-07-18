package middleware

import (
	AppHttp "drto-link/internal/api/http"
	"drto-link/internal/config"
	"errors"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {

		apiKey := config.Get().Link.ApiKey
		requestApiKey := ctx.GetHeader("x-api-key")

		if requestApiKey != apiKey {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, AppHttp.ApiResponse{
				Message: "Unauthorized",
				Error:   errors.New("you are not authorized").Error(),
				Path:    ctx.FullPath(),
			})
			return

		} else {
			ctx.Next()
		}
	}
}
