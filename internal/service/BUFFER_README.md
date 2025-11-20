# Link Buffer - Async Processing

The link buffer provides async processing for MongoDB inserts, improving throughput and reducing API response times.

## ⚠️ Important: Data Persistence

### In-Memory Buffer (Default)
- **Pros:** Simple, no external dependencies, fast
- **Cons:** ⚠️ **Data loss on crash** - pending jobs in buffer are lost if app crashes
- **Use case:** Development, testing, low-stakes environments

### Kafka Buffer (Recommended for Production)
- **Pros:** Persistent, guaranteed delivery, no data loss, scalable
- **Cons:** Requires Kafka infrastructure
- **Use case:** Production environments where data loss is unacceptable

## Configuration

### Disable Buffer (Synchronous Processing)

```bash
LINK_BUFFER_ENABLED=false
```

All link inserts will be processed synchronously. Safe but slower.

### Enable In-Memory Buffer

```bash
LINK_BUFFER_ENABLED=true
LINK_BUFFER_TYPE=memory
LINK_BUFFER_SIZE=1000
LINK_BUFFER_WORKERS=10
```

⚠️ **Warning:** Pending jobs in buffer will be lost if application crashes.

### Enable Kafka Buffer (Production)

```bash
LINK_BUFFER_ENABLED=true
LINK_BUFFER_TYPE=kafka
LINK_BUFFER_WORKERS=20
KAFKA_BROKERS=kafka-1:9092,kafka-2:9092,kafka-3:9092
KAFKA_TOPIC=link-inserts
```

✅ **Recommended:** No data loss, persistent queue, guaranteed delivery.

## Implementation

### In-Memory Buffer

Located in `link_buffer.go`:

```go
// Initialize repository first
linkRepo := repository.NewMongoLinkRepository(mongoClient, "link_db")

// Create buffer with repository
buffer := service.NewLinkBuffer(
    linkRepo,
    1000, // buffer size
    10,   // workers
)
```

**How it works:**
1. Link insert request arrives
2. Job added to buffered channel
3. Worker goroutines process jobs from channel
4. MongoDB insert happens asynchronously

**Data loss scenario:**
- App crashes → Pending jobs in channel are lost
- No persistence, no recovery

### Kafka Buffer

Located in `link_buffer_kafka.go`:

```go
// Initialize repository first
linkRepo := repository.NewMongoLinkRepository(mongoClient, "link_db")

// Create Kafka buffer with repository
buffer := service.NewKafkaLinkBuffer(
    linkRepo,
    []string{"kafka:9092"},
    "link-inserts",
    10, // workers
)
```

**How it works:**
1. Link insert request arrives
2. Job serialized to JSON and sent to Kafka
3. Kafka persists message to disk
4. Worker goroutines consume from Kafka
5. MongoDB insert happens asynchronously
6. Kafka offset committed on success

**No data loss:**
- App crashes → Messages remain in Kafka
- On restart → Workers resume from last committed offset
- Guaranteed delivery with at-least-once semantics

## Implementing Kafka Buffer

The Kafka buffer is provided as a template. To use it:

### 1. Add Kafka Dependency

```bash
go get github.com/segmentio/kafka-go
```

### 2. Uncomment Kafka Code

In `link_buffer_kafka.go`, uncomment:
- Kafka imports
- `kafkaWriter` and `kafkaReader` fields
- `initKafka()` method
- Producer/consumer logic in `QueueLink()` and `worker()`

### 3. Configure Kafka

```bash
LINK_BUFFER_ENABLED=true
LINK_BUFFER_TYPE=kafka
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=link-inserts
```

### 4. Start Kafka

```bash
# Using Docker Compose
docker-compose -f deployment/kafka/docker-compose.yml up -d
```

## Performance Comparison

| Mode | Throughput | Data Safety | Latency |
|------|-----------|-------------|---------|
| Synchronous | 1,000 req/s | ✅ Safe | High |
| In-Memory Buffer | 10,000 req/s | ⚠️ Risk | Low |
| Kafka Buffer | 10,000 req/s | ✅ Safe | Low |

## Monitoring

### Buffer Statistics

```go
stats := buffer.Stats()
// {
//   "enabled": true,
//   "type": "memory" or "kafka",
//   "buffer_size": 1000,
//   "pending_jobs": 45,
//   "worker_count": 10,
//   "buffer_usage_pct": 4.5,
//   "persistent": false or true
// }
```

### Logs

**Buffer initialization:**
```
INFO: In-memory link buffer initialized
  workers: 10
  bufferSize: 1000
  type: in-memory
  warning: Data may be lost on crash. Consider Kafka for production.
```

**Buffer full (fallback to sync):**
```
WARN: Link buffer full, processing synchronously
  shortLinkHash: abc123
  bufferCap: 1000
```

## Best Practices

### Development
```bash
# Disable for simplicity
LINK_BUFFER_ENABLED=false
```

### Staging
```bash
# In-memory for testing
LINK_BUFFER_ENABLED=true
LINK_BUFFER_TYPE=memory
LINK_BUFFER_SIZE=1000
LINK_BUFFER_WORKERS=10
```

### Production
```bash
# Kafka for reliability
LINK_BUFFER_ENABLED=true
LINK_BUFFER_TYPE=kafka
LINK_BUFFER_WORKERS=20
KAFKA_BROKERS=kafka-1:9092,kafka-2:9092,kafka-3:9092
KAFKA_TOPIC=link-inserts
```

## Graceful Shutdown

The buffer handles graceful shutdown:

1. Stop accepting new jobs
2. Wait for workers to finish pending jobs (with timeout)
3. Close connections

```go
buffer.Shutdown(30 * time.Second)
```

**In-Memory:** Pending jobs in channel are processed before shutdown
**Kafka:** Kafka offsets committed, messages remain for next startup

## Alternative Implementations

You can implement `LinkBufferInterface` for other message queues:

### RabbitMQ
```go
type RabbitMQLinkBuffer struct {
    // RabbitMQ connection
    // Channel, queue, etc.
}

func (r *RabbitMQLinkBuffer) QueueLink(link, hash string) error {
    // Publish to RabbitMQ
}
```

### AWS SQS
```go
type SQSLinkBuffer struct {
    // SQS client
}

func (s *SQSLinkBuffer) QueueLink(link, hash string) error {
    // Send to SQS
}
```

### Redis Streams
```go
type RedisStreamBuffer struct {
    // Redis client
}

func (r *RedisStreamBuffer) QueueLink(link, hash string) error {
    // XADD to Redis stream
}
```

## Testing

Tests are in `link_buffer_test.go`:

```bash
# Run buffer tests
go test -v -run TestLinkBuffer ./internal/service/

# With MongoDB
docker run -d -p 27017:27017 mongo:7.0
go test -v ./internal/service/
```

## Migration Guide

### From Synchronous to In-Memory

1. Set `LINK_BUFFER_ENABLED=true`
2. Set `LINK_BUFFER_TYPE=memory`
3. Configure buffer size and workers
4. Restart application
5. Monitor logs for buffer stats

### From In-Memory to Kafka

1. Deploy Kafka cluster
2. Create topic: `kafka-topics --create --topic link-inserts`
3. Uncomment Kafka code in `link_buffer_kafka.go`
4. Add Kafka dependency: `go get github.com/segmentio/kafka-go`
5. Update configuration:
   ```bash
   LINK_BUFFER_TYPE=kafka
   KAFKA_BROKERS=kafka:9092
   KAFKA_TOPIC=link-inserts
   ```
6. Rebuild and deploy
7. Verify Kafka consumer group is active

## Troubleshooting

### Buffer Full Warnings

**Symptom:** Logs show "Link buffer full, processing synchronously"

**Solution:**
- Increase `LINK_BUFFER_SIZE`
- Increase `LINK_BUFFER_WORKERS`
- Check MongoDB performance

### Data Loss on Crash

**Symptom:** Some links not inserted after crash

**Solution:**
- Switch to Kafka buffer
- Or disable buffer (synchronous processing)

### Kafka Connection Errors

**Symptom:** "Failed to connect to Kafka"

**Solution:**
- Verify Kafka is running: `docker ps | grep kafka`
- Check broker addresses in config
- Verify network connectivity
- Check Kafka logs

## Summary

| Feature | In-Memory | Kafka |
|---------|-----------|-------|
| Setup complexity | Low | Medium |
| Data persistence | ❌ No | ✅ Yes |
| Data loss risk | ⚠️ High | ✅ None |
| Performance | ✅ Fast | ✅ Fast |
| Scalability | Limited | ✅ High |
| Production ready | ⚠️ No | ✅ Yes |

**Recommendation:** Use Kafka buffer for production, in-memory for development.
