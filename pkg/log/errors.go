package log

import (
	"github.com/getsentry/sentry-go"
	"intrack-notification/internal/config"
)

// Error logs the error using Zap and sends the error to Sentry.
func Error(section Section, event Event, err error, extra map[ExtraKey]interface{}) {
	NewLogger(config.Get()).Error(section, event, err.Error(), extra)
	sentry.CaptureException(err)
}
