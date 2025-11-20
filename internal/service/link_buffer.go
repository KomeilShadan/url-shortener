package service

import (
	"context"
	"encoding/json"
	"janus/internal/repository"
	"janus/pkg/log"
	"sync"
	"time"
)

// LinkInsertJob represents a link insertion task for async processing.
// Jobs are serializable to JSON for compatibility with message queues like Kafka or RabbitMQ.
type LinkInsertJob struct {
	Link          string    `json:"link"`
	ShortLinkHash string    `json:"short_link_hash"`
	Timestamp     time.Time `json:"timestamp"`
}

// ToJSON serializes the job to JSON for message queue persistence.
func (j *LinkInsertJob) ToJSON() ([]byte, error) {
	return json.Marshal(j)
}

// FromJSON deserializes a job from JSON when consuming from message queues.
func (j *LinkInsertJob) FromJSON(data []byte) error {
	return json.Unmarshal(data, j)
}

// LinkBuffer provides async MongoDB inserts using in-memory buffered channels.
//
// WARNING: Pending jobs in the buffer are lost if the application crashes.
// For production environments requiring data persistence, use KafkaLinkBuffer instead.
//
// The buffer uses a worker pool pattern where multiple goroutines consume jobs
// from a buffered channel and process them concurrently.
type LinkBuffer struct {
	buffer     chan LinkInsertJob
	linkRepo   repository.LinkRepository
	workerPool int
	wg         *sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	enabled    bool
}

var _ LinkBufferInterface = (*LinkBuffer)(nil)

// NewLinkBuffer creates a new in-memory link buffer with the specified configuration.
//
// Parameters:
//   - linkRepo: Link repository for database operations
//   - bufferSize: Channel buffer capacity (recommended: 1000-10000)
//   - workerPool: Number of concurrent workers (recommended: 10-50)
//
// Returns a fully initialized buffer with workers already started.
//
// NOTE: This implementation does not persist pending jobs. For production environments
// where data loss is unacceptable, use NewKafkaLinkBuffer instead.
func NewLinkBuffer(linkRepo repository.LinkRepository, bufferSize, workerPool int) *LinkBuffer {
	ctx, cancel := context.WithCancel(context.Background())
	
	buffer := &LinkBuffer{
		buffer:     make(chan LinkInsertJob, bufferSize),
		linkRepo:   linkRepo,
		workerPool: workerPool,
		ctx:        ctx,
		cancel:     cancel,
		enabled:    true,
		wg:         &sync.WaitGroup{},
	}
	
	buffer.start()
	
	log.Info(log.General, log.Startup, "In-memory link buffer initialized", map[string]interface{}{
		"workers":    workerPool,
		"bufferSize": bufferSize,
		"type":       "in-memory",
		"warning":    "Data may be lost on crash. Consider Kafka for production.",
	})
	
	return buffer
}

// start spawns worker goroutines to process jobs from the buffer.
func (lb *LinkBuffer) start() {
	for i := 0; i < lb.workerPool; i++ {
		lb.wg.Add(1)
		go lb.worker(i)
	}
}

// worker continuously processes jobs from the buffer channel until shutdown.
// Each worker runs in its own goroutine and handles jobs independently.
func (lb *LinkBuffer) worker(id int) {
	defer lb.wg.Done()

	for {
		select {
		case <-lb.ctx.Done():
			return
		case job, ok := <-lb.buffer:
			if !ok {
				return
			}
			lb.processJob(job, id)
		}
	}
}

// processJob performs the database insert operation for a single job.
// Uses repository layer to handle the insert with upsert logic.
func (lb *LinkBuffer) processJob(job LinkInsertJob, workerId int) {
	startTime := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := lb.linkRepo.Insert(ctx, job.Link, job.ShortLinkHash, job.Timestamp)

	if err != nil {
		log.Error(log.Mongodb, log.Insert, err, map[string]interface{}{
			"workerId":      workerId,
			"shortLinkHash": job.ShortLinkHash,
			"duration_ms":   time.Since(startTime).Milliseconds(),
		})
	}
}

// QueueLink adds a link to the buffer for async processing.
//
// If the buffer is full, the job is processed synchronously to prevent blocking
// the client. This ensures the API remains responsive even under heavy load.
//
// Returns nil on success. Errors are logged but not returned to maintain
// non-blocking behavior.
func (lb *LinkBuffer) QueueLink(link, shortLinkHash string) error {
	job := LinkInsertJob{
		Link:          link,
		ShortLinkHash: shortLinkHash,
		Timestamp:     time.Now(),
	}

	select {
	case lb.buffer <- job:
		return nil
	default:
		log.Warn(log.General, log.Insert, "Link buffer full, processing synchronously", map[string]interface{}{
			"shortLinkHash": shortLinkHash,
			"bufferCap":     cap(lb.buffer),
		})
		lb.processJob(job, -1)
		return nil
	}
}

// Shutdown gracefully stops the buffer and waits for pending jobs to complete.
//
// The shutdown process:
//  1. Signals all workers to stop via context cancellation
//  2. Closes the buffer channel (no new jobs accepted)
//  3. Waits for workers to finish processing pending jobs
//  4. Returns when all workers complete or timeout is reached
//
// If timeout is exceeded, some pending jobs may be lost.
func (lb *LinkBuffer) Shutdown(timeout time.Duration) {
	lb.cancel()
	close(lb.buffer)

	done := make(chan struct{})
	go func() {
		lb.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Info(log.General, log.Shutdown, "Link buffer shutdown completed", nil)
		return
	case <-time.After(timeout):
		log.Warn(log.General, log.Shutdown, "Buffer shutdown timeout", map[string]interface{}{
			"timeout_sec": timeout.Seconds(),
		})
	}
}

// Stats returns real-time buffer metrics for monitoring and debugging.
func (lb *LinkBuffer) Stats() map[string]interface{} {
	return map[string]interface{}{
		"enabled":          lb.enabled,
		"type":             "in-memory",
		"buffer_size":      cap(lb.buffer),
		"pending_jobs":     len(lb.buffer),
		"worker_count":     lb.workerPool,
		"buffer_usage_pct": float64(len(lb.buffer)) / float64(cap(lb.buffer)) * 100,
		"persistent":       false,
	}
}

// ProcessJobSync processes a job synchronously, bypassing the buffer.
// Useful for fallback scenarios or when immediate processing is required.
func (lb *LinkBuffer) ProcessJobSync(link, shortLinkHash string) error {
	job := LinkInsertJob{
		Link:          link,
		ShortLinkHash: shortLinkHash,
		Timestamp:     time.Now(),
	}
	lb.processJob(job, -1)
	return nil
}
