package config

type Mongo struct {
	URI      string `env:"MONGO_URI" envDefault:"mongodb://admin:admin@localhost:27717/"`
	Username string `env:"MONGO_USERNAME" envDefault:"admin"`
	Password string `env:"MONGO_PASSWORD" envDefault:"admin"`
	DB       string `env:"MONGO_DB" envDefault:"link"`
}
