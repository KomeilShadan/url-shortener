package main

import (
	"context"
	rest "drto-link/internal/api"
	"drto-link/internal/config"
	"drto-link/pkg/log"
	"drto-link/pkg/mongodb"
	"github.com/caarlos0/env/v11"
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

	rest.InitServer(cfg)

}
