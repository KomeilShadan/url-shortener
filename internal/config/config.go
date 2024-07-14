package config

type Config struct {
	App    App
	Log    Log
	Link   Link
	Mongo  Mongo
	Redis  Redis
	Sentry Sentry
}

var cfg Config

func Get() *Config {
	return &cfg
}
