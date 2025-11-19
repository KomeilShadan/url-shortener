package config

type Mongo struct {
	URI                string `env:"MONGO_URI" envDefault:"mongodb://admin:admin@mongodb:27017/"`
	Username           string `env:"MONGO_USERNAME" envDefault:"admin"`
	Password           string `env:"MONGO_PASSWORD" envDefault:"admin"`
	DB                 string `env:"MONGO_DB" envDefault:"link"`
	ConnectTimeout     int    `env:"MONGO_CONNECT_TIMEOUT" envDefault:"30"`      // seconds
	MaxIdleTime        int    `env:"MONGO_MAX_IDLE_TIME" envDefault:"180"`       // seconds
	MinPoolSize        uint64 `env:"MONGO_MIN_POOL_SIZE" envDefault:"5"`
	MaxPoolSize        uint64 `env:"MONGO_MAX_POOL_SIZE" envDefault:"10"`
	MaxConnecting      uint64 `env:"MONGO_MAX_CONNECTING" envDefault:"10"`
}
