package config

type Redis struct {
	Host         string `env:"REDIS_HOST" envDefault:"127.0.0.1"`
	Port         string `env:"REDIS_PORT" envDefault:"6379"`
	DB           int    `env:"REDIS_DB" envDefault:"0"`
	Password     string `env:"REDIS_PASSWORD" envDefault:""`
	MinIdleConns int    `env:"REDIS_MIN_IDLE_CONNS" envDefault:"0"`
	PoolSize     int    `env:"REDIS_POOL_SIZE" envDefault:"0"`
	PoolTimeout  int    `env:"REDIS_POOL_TIMEOUT" envDefault:"0"`
}
