package middleware

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

func InjectMongoClient(mongo *mongo.Client) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set("mongo", mongo)
		ctx.Next()
	}
}
