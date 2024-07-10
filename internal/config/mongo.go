package config

type Mongo struct {
	URI      string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
	Username string `env:"MONGO_USERNAME" envDefault:"admin"`
	Password string `env:"MONGO_PASSWORD" envDefault:"admin"`
}
