package config

type Sentry struct {
	Dsn         string `env:"SENTRY_DSN"`
	Environment string `env:"SENTRY_ENVIRONMENT" envDefault:"develop"`
}
