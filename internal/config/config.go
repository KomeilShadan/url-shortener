package config

type Config struct {
	App   App
	Log   Log
	Link  Link
	Mongo Mongo
	Redis Redis
}

var cfg Config

func Get() *Config {
	return &cfg
}
