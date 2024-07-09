package config

type Link struct {
	ApiKey string `env:"LINK_API_KEY,required"`
}
