package log

import (
	"github.com/getsentry/sentry-go"
)

// Error logs the error and sends the error to Sentry.
func Error(section Section, event Event, err error, extra map[string]interface{}) {
	extraKey := convertToExtraKeyMap(extra)
	GetLogger().Error(section, event, err.Error(), extraKey)
	sentry.CaptureException(err)
}

// Info logs an informational message.
func Info(section Section, event Event, msg string, extra map[string]interface{}) {
	extraKey := convertToExtraKeyMap(extra)
	GetLogger().Debug(section, event, msg, extraKey)
}

// Debug logs a debug message.
func Debug(section Section, event Event, msg string, extra map[string]interface{}) {
	extraKey := convertToExtraKeyMap(extra)
	GetLogger().Debug(section, event, msg, extraKey)
}

// Warn logs a warning message.
func Warn(section Section, event Event, msg string, extra map[string]interface{}) {
	extraKey := convertToExtraKeyMap(extra)
	GetLogger().Error(section, event, msg, extraKey)
}

// convertToExtraKeyMap converts map[string]interface{} to map[ExtraKey]interface{}
func convertToExtraKeyMap(extra map[string]interface{}) map[ExtraKey]interface{} {
	if extra == nil {
		return nil
	}
	result := make(map[ExtraKey]interface{}, len(extra))
	for k, v := range extra {
		result[ExtraKey(k)] = v
	}
	return result
}
