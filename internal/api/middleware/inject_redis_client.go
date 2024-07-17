package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

func InjectRedisClient(rdb *redis.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("redis", rdb)
		ctx.Next()
	}
}
