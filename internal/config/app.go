package config

type Env string

const (
	local      Env = "local"
	develop    Env = "develop"
	stage      Env = "stage"
	production Env = "production"
)

type App struct {
	Host        string `env:"APP_HOST" envDefault:"localhost"`
	Port        int    `env:"APP_PORT" envDefault:"8080"`
	Key         string `env:"APP_KEY" envDefault:""`
	Mode        string `env:"APP_MODE" envDefault:"debug"`
	Environment Env    `env:"APP_ENV" envDefault:"local"`
}
