package log

import (
	"drto-link/internal/config"
	"go.uber.org/zap"
)

type LoggerInterface interface {
	Init()
	Info(msg string, fields ...zap.Field)
	Debug(sect Section, event Event, msg string, extra map[ExtraKey]interface{})
	Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{})
}

type SyslogLogger struct {
	Logger *zap.Logger
}

func NewLogger(app *config.Config) LoggerInterface {
	if app.Log.Logger == "zap" {
		return ZapLogger(app)
	}

	if app.Log.Logger == "syslog" {
		logger, err := newSyslogLogger()
		if err != nil {
			panic("Failed to initialize syslog logger: " + err.Error())
		}
		return &SyslogLogger{Logger: logger}
	}

	return nil
}
