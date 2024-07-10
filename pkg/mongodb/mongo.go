package mongodb

import (
	"context"
	"drto-link/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

const (
	ConnectTimeout = 30 * time.Second
	MaxIdleTime    = 3 * time.Minute
	MinPoolSize    = 20
	MaxPoolSize    = 300
)

var (
	clientOptions *options.ClientOptions
)

func InitConnection(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {

	clientOptions = options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetAuth(options.Credential{
			Username: cfg.Mongo.Username,
			Password: cfg.Mongo.Password,
		}).
		SetConnectTimeout(ConnectTimeout).
		SetMaxConnIdleTime(MaxIdleTime).
		SetMinPoolSize(MinPoolSize).
		SetMaxPoolSize(MaxPoolSize)

	client, err := mongo.Connect(ctx, clientOptions)

	if err != nil {
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
