package log

import (
	"drto-link/internal/config"
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"log/syslog"
)

func (s *SyslogLogger) Init() {
	// Initialization logic for SyslogLogger if needed
}

func (s *SyslogLogger) Info(msg string, fields ...zap.Field) {
	// Implementation of Info method
	s.writeSyslog(syslog.LOG_INFO, msg)
}

func (s *SyslogLogger) Debug(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogi(sect, event, extra)
	s.Logger.Debug(msg, params...)
}

func (s *SyslogLogger) Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogi(sect, event, extra)
	s.Logger.Error(msg, params...)
}

func prepareLogi(sect Section, event Event, extra map[ExtraKey]interface{}) []zap.Field {
	var fields []zap.Field
	fields = append(fields, zap.String("Section", string(sect)))
	fields = append(fields, zap.String("Event", string(event)))

	for key, value := range extra {
		fields = append(fields, zap.Any(string(key), value))
	}
	return fields
}

func newSyslogLogger() (*zap.Logger, error) {
	cfg := config.Get()
	writer, err := syslog.Dial(cfg.Log.Syslog.Network, cfg.Log.Syslog.Raddr,
		syslog.LOG_WARNING|syslog.LOG_DAEMON, "push-notif")
	if err != nil {
		log.Fatal(err)
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoder := zapcore.NewJSONEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, zapcore.AddSync(writer), zap.NewAtomicLevelAt(zap.InfoLevel))

	return zap.New(core), nil
}

// writeSyslog is an unexported method that writes a log entry to syslog with the specified priority.
func (s *SyslogLogger) writeSyslog(priority syslog.Priority, msg string) {
	writer, err := syslog.New(priority, "push-notification")
	if err != nil {
		fmt.Printf("Failed to initialize syslog: %v", err)
		return
	}
	defer writer.Close()

	_, err = writer.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to write to syslog: %v", err)
	}
}
