package log

import (
	"go.uber.org/zap"
	configs "janus/internal/config"
	"log"
	"sync"
)

var (
	singletonLogger LoggerInterface
	once            sync.Once
)

type LoggerInterface interface {
	Init()
	Info(msg string, fields ...zap.Field)
	Debug(sect Section, event Event, msg string, extra map[ExtraKey]interface{})
	Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{})
}
type SyslogLogger struct {
	*log.Logger // Embed the standard logger
}

func NewLogger(app *configs.Config) LoggerInterface {
	if app.Log.Logger == "zap" {
		return ZapLogger(app)
	} else if app.Log.Logger == "syslog" {
		logger, err := newSyslogLogger()
		if err != nil {
			panic("Failed to initialize syslog logger: " + err.Error())
		}
		return &SyslogLogger{Logger: logger}
	} else if app.Log.Logger == "gelf" {
		logger, err := newGelfLogger(app)
		if err != nil {
			panic("Failed to initialize GELF logger: " + err.Error())
		}
		return logger
	}
	return nil
}

func GetLogger() LoggerInterface {
	once.Do(func() {
		singletonLogger = NewLogger(configs.Get())
		singletonLogger.Init()
	})
	return singletonLogger
}

// prepareLog converts sect and event into zap fields and appends them to the existing fields.
func prepareLogFields(sect Section, event Event, extra map[ExtraKey]interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(extra)+2)
	for k, v := range extra {
		fields = append(fields, zap.Any(string(k), v))
	}
	fields = append(fields, zap.String("section", string(sect)))
	fields = append(fields, zap.String("event", string(event)))
	return fields
}

// prepareLogParams converts sect and event into key-value pairs and appends them to the existing fields.
func prepareLogParams(sect Section, event Event, extra map[ExtraKey]interface{}) []interface{} {
	params := make([]interface{}, 0, len(extra)*2+4)
	for k, v := range extra {
		params = append(params, string(k), v)
	}
	params = append(params, "section", string(sect))
	params = append(params, "event", string(event))
	return params
}
