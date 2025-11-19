package config

type Log struct {
	Logger   string `env:"LOG_LOGGER"`
	Level    string `env:"LOG_LEVEL"`
	FilePath string `env:"LOG_FILE_PATH"`
	Syslog   Syslog
	GrayLog  GrayLog
}

type Syslog struct {
	Network string `env:"SYSLOG_NETWORK"`
	Raddr   string `env:"SYSLOG_RADDR"`
}

type GrayLog struct {
	Raddr    string `env:"GRAYLOG_RADDR"`
	Facility string `env:"GRAYLOG_FACILITY"`
	Protocol string `env:"GRAYLOG_PROTOCOL"`
}
