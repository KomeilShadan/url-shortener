package config

// Link contains configuration for link management and optional async processing.
type Link struct {
	ApiKey string `env:"LINK_API_KEY,required"`

	// Buffer configuration for async processing (disabled by default)
	BufferEnabled         bool   `env:"LINK_BUFFER_ENABLED" envDefault:"false"`
	BufferType            string `env:"LINK_BUFFER_TYPE" envDefault:"memory"`           // Options: memory, kafka
	BufferSize            int    `env:"LINK_BUFFER_SIZE" envDefault:"1000"`             // Only for memory type
	BufferWorkers         int    `env:"LINK_BUFFER_WORKERS" envDefault:"10"`            // Number of worker goroutines
	BufferShutdownTimeout int    `env:"LINK_BUFFER_SHUTDOWN_TIMEOUT" envDefault:"10"`   // Seconds to wait for graceful shutdown

	// Kafka configuration (only used when BufferType=kafka)
	KafkaBrokers string `env:"KAFKA_BROKERS" envDefault:"localhost:9092"`
	KafkaTopic   string `env:"KAFKA_TOPIC" envDefault:"link-inserts"`
}
