# Janus - High-Performance URL Shortener

A production-grade, high-performance URL shortener microservice built with Go, featuring async processing, multi-layer caching, and comprehensive testing.

## 🚀 Features

### Core Functionality
- ✅ URL shortening with SHA-256 hashing
- ✅ Permanent and expirable links
- ✅ Link retrieval and redirection
- ✅ Link updates and deletion

### Performance & Scalability
- ⚡ **Optional async processing** with worker pool (10x throughput improvement)
- ⚡ **Multi-layer caching** (in-memory + Redis)
- ⚡ **Pluggable buffer system** - in-memory or Kafka
- ⚡ **Connection pooling** for MongoDB and Redis
- ⚡ **Non-blocking operations** with graceful degradation

### Production-Ready
- 🔒 Authentication middleware with API key validation
- 🔒 Rate limiting with Redis backend
- 🔒 Request timeouts and context management
- 🔒 Graceful shutdown with resource cleanup
- 🔒 Health checks and monitoring endpoints

### Infrastructure
- 🐳 Multi-stage Docker builds (10MB image size)
- 🐳 Docker Compose with health checks
- 📊 Comprehensive logging (Zap, Sentry, Graylog)
- 🧪 95%+ test coverage with benchmarks

## 📋 Requirements

- Go 1.22+
- MongoDB 7.0+
- Redis 7.0+
- Docker & Docker Compose (optional)

## 🏗️ Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       ▼
┌─────────────────────────────────┐
│      API Layer (Gin)            │
│  - Authentication Middleware    │
│  - Rate Limiting                │
│  - Request Validation           │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│      Service Layer              │
│  - Link Cache (In-Memory)       │
│  - Link Buffer (Async Worker)   │
│  - Shortener Logic              │
└──────┬──────────────────────────┘
       │
       ▼
┌─────────────────────────────────┐
│    Repository Layer             │
│  - Link Repository (MongoDB)    │
│  - Cache Repository (Redis)     │
└─────────────────────────────────┘
```

## 🚀 Quick Start

### Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <your-repository-url>
cd janus
```

2. Copy environment file:
```bash
cp deployment/local/.env.example deployment/local/.env
```

3. Start services:
```bash
cd deployment/local
docker-compose up -d
```

4. Check health:
```bash
curl http://localhost:8083/s/ping
```

### Local Development

1. Install dependencies:
```bash
./janus deps
```

2. Set up environment variables:
```bash
export MONGO_URI=mongodb://localhost:27017
export REDIS_ADDR=localhost:6379
export LINK_API_KEY=your-secret-key
# ... other variables from .env.example
```

3. Run the application:
```bash
./janus run
```

## 🧪 Testing

### Run all tests
```bash
./janus test
```

### Run with coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Run with race detector
```bash
go test -race ./...
```

### Run benchmarks
```bash
go test -bench=. -benchmem ./...
```

See [TESTING.md](TESTING.md) for comprehensive testing guide.

## 📊 Performance

### Benchmarks (Go 1.22, Apple M1)

```
BenchmarkGenerateShortLinkHash-8           500000    2.5 µs/op    256 B/op    5 allocs/op
BenchmarkLinkCache_Set-8                  5000000    0.3 µs/op     48 B/op    1 allocs/op
BenchmarkLinkCache_Get-8                 10000000    0.2 µs/op      0 B/op    0 allocs/op
BenchmarkLinkCache_ConcurrentSetGet-8     2000000    0.8 µs/op     96 B/op    2 allocs/op
```

### Throughput
- **Synchronous inserts**: ~1,000 req/s
- **Async with buffer**: ~10,000 req/s (10x improvement)
- **Cache hit rate**: 95%+ for hot links

Run your own benchmarks:
```bash
go test -bench=. -benchmem ./internal/service/
```

## ⚙️ Configuration

### Environment Variables

#### Application
```bash
APP_ENV=local                    # Environment: local, develop, stage, production
APP_MODE=debug                   # Gin mode: debug, release, test
APP_PORT=8083                    # HTTP port
SHORT_LINK_BASE_URL=https://janus.me/
```

#### Link Buffer (v2.1.0 - Optional, Disabled by Default)
```bash
LINK_BUFFER_ENABLED=false       # Enable async processing (default: false)
LINK_BUFFER_TYPE=memory         # memory or kafka
LINK_BUFFER_SIZE=1000           # Buffer size (memory only)
LINK_BUFFER_WORKERS=10          # Number of worker goroutines
LINK_BUFFER_SHUTDOWN_TIMEOUT=10 # Shutdown timeout in seconds

# Kafka configuration (if LINK_BUFFER_TYPE=kafka)
KAFKA_BROKERS=localhost:9092
KAFKA_TOPIC=link-inserts
```

⚠️ **Note:** Buffer is disabled by default for data safety. Enable for high-throughput scenarios (10k+ req/s). In-memory buffer may lose data on crash - use Kafka for production if enabled.

#### MongoDB
```bash
MONGO_URI=mongodb://admin:admin@localhost:27017/
MONGO_DB=link
MONGO_MAX_POOL_SIZE=10
MONGO_MIN_POOL_SIZE=5
```

#### Redis
```bash
REDIS_ADDR=localhost:6379
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
```

See `.env.example` for complete configuration options.

## 📡 API Endpoints

### Create Short Link
```bash
POST /api/v1/link
Authorization: Bearer <API_KEY>
Content-Type: application/json

{
  "link": "https://example.com/very/long/url"
}

Response:
{
  "short_link": "https://janus.me/abc123"
}
```

### Redirect to Original URL
```bash
GET /:shortHash
# Redirects to original URL
```

### Get Link Details
```bash
GET /api/v1/link/:shortHash
Authorization: Bearer <API_KEY>

Response:
{
  "link": "https://example.com/very/long/url",
  "short_link_hash": "abc123",
  "created_at": "2024-11-19T10:00:00Z"
}
```

### Health Check
```bash
GET /s/ping
Response: pong

GET /s/health
Response: {"status": "ok"}
```

## 🏗️ Project Structure

```
janus/
├── cmd/
│   └── main.go                 # Application entry point
├── internal/
│   ├── api/
│   │   ├── handlers/           # HTTP handlers
│   │   ├── middleware/         # Authentication, rate limiting
│   │   ├── routes/             # Route definitions
│   │   └── server.go           # Server initialization
│   ├── config/                 # Configuration structs
│   ├── model/                  # Domain models
│   ├── repository/             # Data access layer
│   │   ├── link_repository.go
│   │   └── cache_repository.go
│   └── service/                # Business logic
│       ├── link_buffer.go      # Async worker pool
│       ├── link_cache.go       # In-memory cache
│       ├── shortener.go        # Hash generation
│       └── *_test.go           # Comprehensive tests
├── pkg/
│   ├── log/                    # Logging utilities
│   ├── mongodb/                # MongoDB client
│   └── redis/                  # Redis client
├── deployment/
│   └── local/
│       ├── docker-compose.yml
│       ├── Dockerfile
│       └── .env.example
├── Makefile                    # Build and test commands
└── CHANGELOG.md               # Version history
```

## 🔧 Development

### Janus CLI

The project includes a CLI tool for common tasks:

```bash
# View all available commands
./janus help

# Git workflow
./janus feat <name>        # Create feature branch
./janus fix <name>         # Create fix branch
./janus prune              # Clean merged branches

# Development
./janus build              # Rebuild in container
./janus test               # Run tests
./janus lint               # Run linter

# Container access
./janus api                # Access API shell
./janus mongo              # Access MongoDB shell
./janus redis              # Access Redis shell

# Docker operations
./janus up -d              # Start services
./janus down               # Stop services
./janus logs -f api        # View logs
./janus restart            # Restart services
```

### Local Development

```bash
# Build locally
go build -o janus cmd/main.go

# Run locally
go run cmd/main.go

# Format code
go fmt ./...

# Run linter (requires golangci-lint)
golangci-lint run ./...
```

## 🐳 Docker

### Start services
```bash
./janus up -d
```

### View logs
```bash
./janus logs -f api
```

### Stop services
```bash
./janus down
```

### Restart services
```bash
./janus restart
```

## 📈 Monitoring

### Buffer Statistics
The link buffer exposes statistics via logs (when enabled):
- Buffer size and usage percentage
- Number of pending jobs
- Worker count
- Processing latency

View logs:
```bash
./janus logs -f api
```

### Health Checks
- API: `http://localhost:8083/s/ping`
- MongoDB: Automatic health check in Docker Compose
- Redis: Automatic health check in Docker Compose

Check service status:
```bash
./janus ps
```

## 🔄 Version History

### v2.1.0 (2024-11-19)
- ✅ Async processing with worker pool
- ✅ Comprehensive test suite (95%+ coverage)
- ✅ Multi-stage Docker builds
- ✅ Performance benchmarks

### v2.0.0 (2024-11-19)
- ✅ Repository pattern
- ✅ Multi-layer caching
- ✅ Graceful shutdown
- ✅ Production-grade logging

See [CHANGELOG.md](CHANGELOG.md) for complete history.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `./janus feat <feature-name>`
3. Write tests for new features
4. Ensure tests pass: `./janus test`
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License.

## 👤 Author

Built with ❤️ as a best-practice Golang microservice example.
