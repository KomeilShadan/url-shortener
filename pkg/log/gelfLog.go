package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	configs "janus/internal/config"
	"log"
	"os"
)

type ZapLoggerWrapper struct {
	*zap.SugaredLogger
}

func (z ZapLoggerWrapper) Init() {
	// TODO implement me
	panic("implement me")
}

func (z ZapLoggerWrapper) Info(msg string, fields ...zap.Field) {
	z.SugaredLogger.Desugar().Info(msg, fields...)
}

func (z ZapLoggerWrapper) Debug(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	fields := prepareLogFields(sect, event, extra)
	z.SugaredLogger.Desugar().Debug(msg, fields...)
}

func (z ZapLoggerWrapper) Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	fields := prepareLogFields(sect, event, extra)
	// Log the error message
	z.SugaredLogger.Desugar().Error(msg, fields...)

	// Check if the log message was successfully sent
	if err := z.SugaredLogger.Sync(); err != nil {
		fmt.Println("Failed to sync logger:", err)
	}
}

func newGelfLogger(app *configs.Config) (LoggerInterface, error) {
	host, err := os.Hostname()
	if err != nil {
		log.Printf("Failed to get hostname: %v", err)
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}

	protocol := app.Log.GrayLog.Protocol // Add this line to get the protocol from the config
	gelfCore, err := newGelfCore(zapcore.DebugLevel, protocol, app.Log.GrayLog.Raddr, host, app.Log.GrayLog.Facility)
	if err != nil {
		log.Printf("Failed to create GELF core: %v", err)
		return nil, fmt.Errorf("failed to create GELF core: %w", err)
	}

	zlogger := zap.New(gelfCore, zap.Development(), zap.AddCaller(),
		zap.AddStacktrace(zap.NewAtomicLevelAt(zap.ErrorLevel)),
	)
	defer func() {
		if err := zlogger.Sync(); err != nil {
			log.Printf("Failed to sync logger: %v", err)
		}
	}()

	return &ZapLoggerWrapper{SugaredLogger: zlogger.Sugar()}, nil
}

func newGelfCore(level zapcore.Level, protocol string, addr string, host string, facility string) (zapcore.Core, error) {
	// Create a new GELF core using the zap-gelf package
	gelfCore, err := NewCore(
		Addr(addr),
		Host(host),
		Level(level),
		MessageKey("short_message"),
		StacktraceKey("full_message"),
		Version("1.1"),
	)
	if err != nil {
		log.Printf("Failed to create GELF core: %v", err)
		return nil, fmt.Errorf("failed to create GELF core: %w", err)
	}

	// Add the facility field to the core
	gelfCore = gelfCore.With([]zapcore.Field{zap.String("facility", facility)})

	return gelfCore, nil
}
