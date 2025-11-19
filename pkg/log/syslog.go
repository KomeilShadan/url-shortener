package log

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	configs "janus/internal/config"
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
	// Convert the structured data (params) into a JSON string for logging.
	// This is one way to handle structured data; adjust according to your needs.
	extraData, err := json.Marshal(params)
	if err != nil {
		// If marshaling fails, log the error message without structured data.
		s.Logger.Print("Error logging structured data: ", err)
		s.Logger.Print(msg)
		return
	}

	// Log the original message along with the structured data as a JSON string.
	s.Logger.Printf("Error: %s | Data: %s", msg, string(extraData))
}

func (s *SyslogLogger) Error(sect Section, event Event, msg string, extra map[ExtraKey]interface{}) {
	params := prepareLogi(sect, event, extra)
	// Convert the structured data (params) into a JSON string for logging.
	// This is one way to handle structured data; adjust according to your needs.
	extraData, err := json.Marshal(params)
	if err != nil {
		// If marshaling fails, log the error message without structured data.
		s.Logger.Print("Error logging structured data: ", err)
		s.Logger.Print(msg)
		return
	}
	//s.Logger.SetFlags(0)
	// Log the original message along with the structured data as a JSON string.
	s.Logger.Printf("Error: %s | Data: %s", msg, string(extraData))
}

func prepareLogi(sect Section, event Event, extra map[ExtraKey]interface{}) []zap.Field {
	var fields []zap.Field
	fields = append(fields, zap.String("Section", string(sect)))
	fields = append(fields, zap.String("Event", string(event)))
	fields = append(fields, zap.String("Tag", "janus"))

	for key, value := range extra {
		fields = append(fields, zap.Any(string(key), value))
	}
	return fields
}

func newSyslogLogger() (*log.Logger, error) {
	cfg := configs.Get()
	writer, err := syslog.Dial(cfg.Log.Syslog.Network, cfg.Log.Syslog.Raddr,
		syslog.LOG_WARNING|syslog.LOG_DAEMON, "janus")

	if err != nil {
		log.Fatal(err)
	}
	//	log.SetFlags(0)
	return log.New(writer, "", 0), nil
}

// writeSyslog is an unexported method that writes a log entry to syslog with the specified priority.
func (s *SyslogLogger) writeSyslog(priority syslog.Priority, msg string) {
	writer, err := syslog.New(priority, "janus")
	if err != nil {
		fmt.Printf("Failed to initialize syslog: %v", err)
		return
	}

	defer func(writer *syslog.Writer) { _ = writer.Close() }(writer)

	_, err = writer.Write([]byte(msg))
	if err != nil {
		fmt.Printf("Failed to write to syslog: %v", err)
	}
}
