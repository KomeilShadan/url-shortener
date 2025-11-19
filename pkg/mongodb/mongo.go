package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"janus/internal/config"
	"janus/pkg/log"
	"time"
)

var (
	clientOptions *options.ClientOptions
)

func InitConnection(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	connectTimeout := time.Duration(cfg.Mongo.ConnectTimeout) * time.Second
	maxIdleTime := time.Duration(cfg.Mongo.MaxIdleTime) * time.Second

	clientOptions = options.Client().
		ApplyURI(cfg.Mongo.URI).
		//SetAuth(options.Credential{
		//	Username: cfg.Mongo.Username,
		//	Password: cfg.Mongo.Password,
		//}).
		SetConnectTimeout(connectTimeout).
		SetMaxConnIdleTime(maxIdleTime).
		SetMinPoolSize(cfg.Mongo.MinPoolSize).
		SetMaxPoolSize(cfg.Mongo.MaxPoolSize).
		SetMaxConnecting(cfg.Mongo.MaxConnecting)

	log.Info(log.Mongodb, log.Startup, "Initializing MongoDB connection", map[string]interface{}{
		"uri":             cfg.Mongo.URI,
		"min_pool_size":   cfg.Mongo.MinPoolSize,
		"max_pool_size":   cfg.Mongo.MaxPoolSize,
		"max_connecting":  cfg.Mongo.MaxConnecting,
		"connect_timeout": cfg.Mongo.ConnectTimeout,
		"max_idle_time":   cfg.Mongo.MaxIdleTime,
	})

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		log.Error(log.Mongodb, log.Startup, err, map[string]interface{}{
			"uri": cfg.Mongo.URI,
		})
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		log.Error(log.Mongodb, log.Startup, err, map[string]interface{}{
			"uri":    cfg.Mongo.URI,
			"reason": "ping failed",
		})
		return nil, err
	}

	log.Info(log.Mongodb, log.Startup, "MongoDB connection established successfully", map[string]interface{}{
		"uri": cfg.Mongo.URI,
	})

	return client, nil
}
