package main

import (
	"context"
	rest "drto-link/internal/api"
	"drto-link/internal/config"
	"drto-link/pkg/log"
	"drto-link/pkg/mongodb"
	"drto-link/pkg/redis"
	"github.com/caarlos0/env/v11"
	"github.com/getsentry/sentry-go"
	"os"
	"time"
)

var (
	cfg *config.Config
)

func main() {

	cfg = config.Get()
	if err := env.Parse(cfg); err != nil {
		log.Error(log.Config, log.Startup, err, nil)
		os.Exit(1)
	}

	Logger := log.NewLogger(cfg)
	// Initialize the new Logger
	Logger.Init()
	// Log messages using the Logger
	Logger.Error(log.Internal, log.Startup, "This is an error message", nil)

	//initial configuration
	err := sentry.Init(sentry.ClientOptions{
		Dsn:         cfg.Sentry.Dsn,
		Environment: cfg.Sentry.Environment,
	})
	if err != nil {
		log.Error(log.Sentry, log.Startup, err, nil)
		os.Exit(1)
	}
	defer sentry.Flush(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	mongo, err := mongodb.InitConnection(ctx, cfg)
	if err != nil {
		log.Error(log.Mongodb, log.Startup, err, nil)
		os.Exit(1)
	}
	defer func() {
		if err := mongo.Disconnect(ctx); err != nil {
			log.Error(log.Mongodb, log.Startup, err, nil)
			os.Exit(1)
		}
	}()

	rdb := redis.InitConnection(cfg)
	defer func() {
		err := rdb.Close()
		if err != nil {

		}
	}()

	rest.InitServer(cfg, rdb, mongo)

}
