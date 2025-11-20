package service

import "time"

// LinkBufferInterface defines the contract for async link processing implementations.
//
// This interface allows pluggable buffer backends:
//   - LinkBuffer: In-memory channels (fast, but data loss on crash)
//   - KafkaLinkBuffer: Kafka-based (persistent, no data loss)
//   - NoOpBuffer: Synchronous processing (no async, safe)
//   - Custom implementations: RabbitMQ, AWS SQS, Redis Streams, etc.
type LinkBufferInterface interface {
	// QueueLink adds a link to the buffer for async processing.
	// Returns nil on success, error on failure.
	QueueLink(link, shortLinkHash string) error

	// Shutdown gracefully stops the buffer and waits for pending jobs.
	// The timeout parameter specifies maximum wait time for completion.
	Shutdown(timeout time.Duration)

	// Stats returns real-time buffer metrics for monitoring.
	Stats() map[string]interface{}
}

// NoOpBuffer is a null implementation that performs no async processing.
// When used, the caller must handle link insertion synchronously.
// This is the default when LINK_BUFFER_ENABLED=false.
type NoOpBuffer struct{}

func (n *NoOpBuffer) QueueLink(link, shortLinkHash string) error {
	return nil
}

func (n *NoOpBuffer) Shutdown(timeout time.Duration) {}

func (n *NoOpBuffer) Stats() map[string]interface{} {
	return map[string]interface{}{
		"enabled": false,
		"type":    "noop",
	}
}
