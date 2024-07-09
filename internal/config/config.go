package config

type Config struct {
	App  App
	Log  Log
	Link Link
}

var cfg Config

func Get() *Config {
	return &cfg
}
