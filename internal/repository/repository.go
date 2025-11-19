package repository

import (
	"github.com/go-redis/redis/v8"
	"go.mongodb.org/mongo-driver/mongo"
	"sync"
)

// Container holds all repository instances
type Container struct {
	Link  LinkRepository
	Cache CacheRepository
}

var (
	container *Container
	once      sync.Once
)

// InitRepositories initializes all repositories
func InitRepositories(mongoClient *mongo.Client, redisClient *redis.Client, dbName string) *Container {
	once.Do(func() {
		container = &Container{
			Link: NewMongoLinkRepository(mongoClient, dbName),
		}
		// Initialize Redis cache repository only if Redis client is provided
		if redisClient != nil {
			container.Cache = NewRedisCacheRepository(redisClient)
		}
	})
	return container
}

// GetRepositories returns the repository container
func GetRepositories() *Container {
	return container
}
