package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"intrack-notification/internal/config"
	"sync"
	"time"
)

var once sync.Once
var zapSingletonLogger *zap.SugaredLogger

type Zap struct {
	app    *config.Config
	logger *zap.SugaredLogger
}

func (z *Zap) Info(msg string, fields ...zap.Field) {

}

func (z *Zap) Debug(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLog(sect, event, extra)
	z.logger.Debugw(msg, params...)
}

func (z *Zap) Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLog(sect, event, extra)
	z.logger.Errorw(msg, params...)
}

func ZapLogger(app *config.Config) *Zap {
	z := &Zap{app: app}
	z.Init()
	return z
}
func (z *Zap) Init() {
	once.Do(func() {
		path := fmt.Sprintf("%s/%s.log", z.app.Log.FilePath, time.Now().Format(time.RFC3339))
		logOutput := &lumberjack.Logger{
			Filename:   path, // Change this to the desired file path
			MaxSize:    1,    // Megabytes
			MaxAge:     5,    // Days
			LocalTime:  true,
			MaxBackups: 10,
			Compress:   true,
		}
		encoderConfig := zap.NewProductionEncoderConfig()
		zapcore.Level.Enabled(zapcore.ErrorLevel, zapcore.DebugLevel)
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(logOutput),
			zap.NewAtomicLevelAt(zap.DebugLevel),
		)

		logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel)).
			Sugar().
			With("appName", "myApp", "logger", "zap")
		zapSingletonLogger = logger
	})
	z.logger = zapSingletonLogger
	z.logger.Errorw("message", "key", "value")
}

func prepareLog(sect Section, event Event, extra map[ExtraKey]interface{}) []interface{} {
	if extra == nil {
		extra = make(map[ExtraKey]interface{})
	}
	extra["Section"] = sect
	extra["Event"] = event
	params := MapToZapParams(extra)
	fmt.Println(params)
	return params
}
