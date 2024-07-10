package config

type Redis struct {
	Host string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	Port string `env:"REDIS_PORT" envDefault:"6379"`
}
