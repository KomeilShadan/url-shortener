package config

type Log struct {
	Logger   string `env:"LOG_LOGGER"`
	Level    string `env:"LOG_LEVEL"`
	FilePath string `env:"LOG_FILE_PATH"`
	Syslog   Syslog
}

type Syslog struct {
	Network string `env:"SYSLOG_NETWORK"`
	Raddr   string `env:"SYSLOG_RADDR"`
}
