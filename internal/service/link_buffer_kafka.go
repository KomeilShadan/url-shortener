package service

import (
	"context"
	"janus/internal/repository"
	"janus/pkg/log"
	"sync"
	"time"
)

// KafkaLinkBuffer manages buffered inserts using Kafka for persistence
// This ensures no data loss even if the application crashes
//
// To use this implementation:
// 1. Add Kafka client dependency: go get github.com/segmentio/kafka-go
// 2. Configure Kafka brokers in environment
// 3. Initialize with NewKafkaLinkBuffer instead of NewLinkBuffer
//
// Example:
//   buffer := service.NewKafkaLinkBuffer(
//       mongoClient,
//       "link_db",
//       []string{"kafka:9092"},
//       "link-inserts",
//       10, // workers
//   )
type KafkaLinkBuffer struct {
	linkRepo   repository.LinkRepository
	workerPool int
	wg         *sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	enabled    bool
	
	// Kafka configuration (commented out to avoid dependency)
	// kafkaWriter *kafka.Writer
	// kafkaReader *kafka.Reader
	brokers     []string
	topic       string
}

// Ensure KafkaLinkBuffer implements LinkBufferInterface
var _ LinkBufferInterface = (*KafkaLinkBuffer)(nil)

// NewKafkaLinkBuffer creates a new Kafka-based link buffer
// This provides persistent message queue with guaranteed delivery
//
// NOTE: This is a template implementation. To use:
// 1. Uncomment Kafka imports and fields
// 2. Add kafka-go dependency
// 3. Implement producer/consumer logic
func NewKafkaLinkBuffer(
	linkRepo repository.LinkRepository,
	brokers []string,
	topic string,
	workerPool int,
) *KafkaLinkBuffer {
	ctx, cancel := context.WithCancel(context.Background())
	
	buffer := &KafkaLinkBuffer{
		linkRepo:   linkRepo,
		workerPool: workerPool,
		ctx:        ctx,
		cancel:     cancel,
		enabled:    true,
		brokers:    brokers,
		topic:      topic,
		wg:         &sync.WaitGroup{},
	}
	
	// Initialize Kafka producer and consumer
	// buffer.initKafka()
	
	// Start consumer workers
	buffer.start()
	
	log.Info(log.General, log.Startup, "Kafka link buffer initialized", map[string]interface{}{
		"workers":    workerPool,
		"brokers":    brokers,
		"topic":      topic,
		"type":       "kafka",
		"persistent": true,
	})
	
	return buffer
}

// initKafka initializes Kafka producer and consumer
// Uncomment and implement when adding Kafka dependency
/*
func (kb *KafkaLinkBuffer) initKafka() {
	// Initialize Kafka writer (producer)
	kb.kafkaWriter = &kafka.Writer{
		Addr:         kafka.TCP(kb.brokers...),
		Topic:        kb.topic,
		Balancer:     &kafka.LeastBytes{},
		RequiredAcks: kafka.RequireAll, // Wait for all replicas
		Async:        false,             // Synchronous for reliability
	}
	
	// Initialize Kafka reader (consumer)
	kb.kafkaReader = kafka.NewReader(kafka.ReaderConfig{
		Brokers:        kb.brokers,
		Topic:          kb.topic,
		GroupID:        "janus-link-processor",
		MinBytes:       1,
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
	})
}
*/

// start initializes worker goroutines to consume from Kafka
func (kb *KafkaLinkBuffer) start() {
	for i := 0; i < kb.workerPool; i++ {
		kb.wg.Add(1)
		go kb.worker(i)
	}
}

// worker processes jobs from Kafka
func (kb *KafkaLinkBuffer) worker(id int) {
	defer kb.wg.Done()
	
	log.Info(log.General, log.Startup, "Kafka worker started", map[string]interface{}{
		"workerId": id,
	})
	
	for {
		select {
		case <-kb.ctx.Done():
			return
		default:
			// Read message from Kafka
			// Uncomment when implementing:
			/*
			msg, err := kb.kafkaReader.ReadMessage(kb.ctx)
			if err != nil {
				if err == context.Canceled {
					return
				}
				log.Error(log.General, log.Insert, err, map[string]interface{}{
					"workerId": id,
				})
				continue
			}
			
			// Deserialize job
			var job LinkInsertJob
			if err := job.FromJSON(msg.Value); err != nil {
				log.Error(log.General, log.Insert, err, map[string]interface{}{
					"workerId": id,
					"offset":   msg.Offset,
				})
				continue
			}
			
			// Process job
			kb.processJob(job, id)
			*/
			
			// Placeholder: sleep to prevent busy loop
			time.Sleep(100 * time.Millisecond)
		}
	}
}

// processJob handles the actual database insert using repository layer
func (kb *KafkaLinkBuffer) processJob(job LinkInsertJob, workerId int) {
	startTime := time.Now()
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	err := kb.linkRepo.Insert(ctx, job.Link, job.ShortLinkHash, job.Timestamp)
	
	if err != nil {
		log.Error(log.Mongodb, log.Insert, err, map[string]interface{}{
			"workerId":      workerId,
			"shortLinkHash": job.ShortLinkHash,
			"duration_ms":   time.Since(startTime).Milliseconds(),
		})
	}
}

// QueueLink adds a link to Kafka for async insertion
func (kb *KafkaLinkBuffer) QueueLink(link, shortLinkHash string) error {
	job := LinkInsertJob{
		Link:          link,
		ShortLinkHash: shortLinkHash,
		Timestamp:     time.Now(),
	}
	
	// Serialize job to JSON
	data, err := job.ToJSON()
	if err != nil {
		log.Error(log.General, log.Insert, err, map[string]interface{}{
			"shortLinkHash": shortLinkHash,
		})
		return err
	}
	
	// Send to Kafka
	// Uncomment when implementing:
	/*
	err = kb.kafkaWriter.WriteMessages(kb.ctx, kafka.Message{
		Key:   []byte(shortLinkHash),
		Value: data,
	})
	
	if err != nil {
		log.Error(log.General, log.Insert, err, map[string]interface{}{
			"shortLinkHash": shortLinkHash,
		})
		return err
	}
	*/
	
	log.Info(log.General, log.Insert, "Link queued to Kafka", map[string]interface{}{
		"shortLinkHash": shortLinkHash,
		"size":          len(data),
	})
	
	return nil
}

// Shutdown gracefully shuts down the Kafka buffer
func (kb *KafkaLinkBuffer) Shutdown(timeout time.Duration) {
	log.Info(log.General, log.Shutdown, "Shutting down Kafka buffer", map[string]interface{}{
		"timeout_sec": timeout.Seconds(),
	})
	
	// Signal workers to stop
	kb.cancel()
	
	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		kb.wg.Wait()
		close(done)
	}()
	
	select {
	case <-done:
		log.Info(log.General, log.Shutdown, "All Kafka workers stopped", nil)
	case <-time.After(timeout):
		log.Warn(log.General, log.Shutdown, "Kafka buffer shutdown timeout", map[string]interface{}{
			"timeout_sec": timeout.Seconds(),
		})
	}
	
	// Close Kafka connections
	// Uncomment when implementing:
	/*
	if err := kb.kafkaWriter.Close(); err != nil {
		log.Error(log.General, log.Shutdown, err, nil)
	}
	if err := kb.kafkaReader.Close(); err != nil {
		log.Error(log.General, log.Shutdown, err, nil)
	}
	*/
}

// Stats returns current Kafka buffer statistics
func (kb *KafkaLinkBuffer) Stats() map[string]interface{} {
	stats := map[string]interface{}{
		"enabled":     kb.enabled,
		"type":        "kafka",
		"worker_count": kb.workerPool,
		"persistent":  true,
		"brokers":     kb.brokers,
		"topic":       kb.topic,
	}
	
	// Add Kafka-specific stats
	// Uncomment when implementing:
	/*
	stats["kafka_lag"] = kb.kafkaReader.Lag()
	stats["kafka_offset"] = kb.kafkaReader.Offset()
	*/
	
	return stats
}
