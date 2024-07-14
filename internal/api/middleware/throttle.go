package middleware

import (
	AppHttp "drto-link/internal/api/http"
	"drto-link/pkg/log"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"
	"strconv"
	"time"
)

func Throttle(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// everytime a user queries, check if the IP is already in database,
		// if yes, decrement the calls remaining by one, else add the IP to database
		// with expiry of `30mins`. So in this case the user will be able to send 10
		// requests every 30 minutes
		val, err := rdb.Get(rdb.Context(), ctx.ClientIP()).Result()
		if errors.Is(err, redis.Nil) {
			err = rdb.Set(rdb.Context(), ctx.ClientIP(), os.Getenv("API_QUOTA"), 30*time.Minute).Err()
			if err != nil {
				log.Error(log.Redis, log.Insert, err, nil)
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			err = rdb.Decr(rdb.Context(), ctx.ClientIP()).Err()
			if err != nil {
				log.Error(log.Redis, log.Insert, err, nil)
				ctx.AbortWithStatus(http.StatusInternalServerError)
			}
			defer ctx.Next()
		} else {
			intval, _ := strconv.Atoi(val)
			if intval <= 0 {
				limit, _ := rdb.TTL(rdb.Context(), ctx.ClientIP()).Result()

				ctx.AbortWithStatusJSON(http.StatusForbidden, AppHttp.ApiResponse{
					Message: "Forbidden",
					Error:   errors.New("rate limit exceeded").Error(),
					Data: map[string]any{
						"rate_limit_reset": limit / time.Nanosecond / time.Minute,
					},
					Path: ctx.FullPath(),
				})
			}
		}
	}
}
