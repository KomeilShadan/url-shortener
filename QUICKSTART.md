# Janus Quick Start Guide

Get up and running with Janus in 5 minutes.

## 🚀 Installation

### Option 1: Docker Compose (Recommended)

```bash
# Clone repository
git clone <your-repository-url>
cd janus

# Setup environment
cp deployment/local/.env.example deployment/local/.env

# Start services
./janus up -d

# Verify
curl http://localhost:8083/s/ping
# Expected: pong
```

### Option 2: Local Development

```bash
# Prerequisites: Go 1.22+, MongoDB, Redis

# Install dependencies
go mod download

# Set environment variables
export MONGO_URI=mongodb://admin:admin@localhost:27017/
export REDIS_ADDR=localhost:6379
export LINK_API_KEY=secret-api-key
export MONGO_DB=link
export APP_PORT=8083
export SHORT_LINK_BASE_URL=http://localhost:8083/

# Run
go run cmd/main.go
```

## 📝 Basic Usage

### 1. Create Short Link

```bash
curl -X POST http://localhost:8083/api/v1/link \
  -H "Authorization: Bearer secret-api-key" \
  -H "Content-Type: application/json" \
  -d '{"link": "https://example.com/very/long/url"}'
```

**Response:**
```json
{
  "short_link": "http://localhost:8083/abc123"
}
```

### 2. Use Short Link

```bash
# Redirects to original URL
curl -L http://localhost:8083/abc123
```

### 3. Get Link Details

```bash
curl http://localhost:8083/api/v1/link/abc123 \
  -H "Authorization: Bearer secret-api-key"
```

**Response:**
```json
{
  "link": "https://example.com/very/long/url",
  "short_link_hash": "abc123",
  "created_at": "2024-11-19T10:00:00Z"
}
```

## 🧪 Testing

```bash
# Run all tests
./janus test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run benchmarks
go test -bench=. -benchmem ./...
```

## 🔧 Configuration

### Essential Environment Variables

```bash
# Application
APP_PORT=8083
LINK_API_KEY=your-secret-key

# Buffer (v2.1.0 - Optional, disabled by default)
LINK_BUFFER_ENABLED=false  # Set to true for async processing
LINK_BUFFER_TYPE=memory    # memory or kafka
LINK_BUFFER_SIZE=1000      # Buffer size (memory only)
LINK_BUFFER_WORKERS=10     # Worker count

# MongoDB
MONGO_URI=mongodb://admin:admin@localhost:27017/
MONGO_DB=link

# Redis
REDIS_ADDR=localhost:6379
```

⚠️ **Note:** Buffer is disabled by default. Enable for high-throughput scenarios.

See `.env.example` for all options.

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:8083/s/ping
# Response: pong
```

### View Logs

```bash
# Docker
./janus logs -f api

# All services
./janus logs -f

# Local
# Logs to stdout
```

### Buffer Statistics

Check logs for buffer stats:
```json
{
  "buffer_size": 1000,
  "pending_jobs": 45,
  "worker_count": 10,
  "buffer_usage_pct": 4.5
}
```

## 🐳 Docker Commands

```bash
# Start services
./janus up -d

# Stop services
./janus down

# View logs
./janus logs -f api

# Restart service
./janus restart api

# Rebuild
./janus build --no-cache

# Access containers
./janus api      # API container shell
./janus mongo    # MongoDB shell
./janus redis    # Redis shell
```

## 🔍 Troubleshooting

### Service won't start

```bash
# Check logs
./janus logs api

# Check dependencies
./janus ps
```

### Can't connect to MongoDB

```bash
# Check MongoDB
./janus mongo
# Then run: db.adminCommand('ping')

# Restart MongoDB
./janus restart mongodb
```

### API returns 401

```bash
# Check API key in .env
cat deployment/local/.env | grep LINK_API_KEY

# Use correct key in request
curl -H "Authorization: Bearer YOUR_KEY" ...
```

## 📚 Next Steps

- Read [README.md](../README.md) for detailed documentation
- Check [TESTING.md](../TESTING.md) for testing guide
- Review [CHANGELOG.md](../CHANGELOG.md) for version history
- Run `./janus help` to see all CLI commands

## 🎯 Performance Tips

### High Throughput (10k+ req/s)

Edit `deployment/local/.env`:
```bash
# Enable async buffer
LINK_BUFFER_ENABLED=true
LINK_BUFFER_SIZE=5000
LINK_BUFFER_WORKERS=20

# Increase connection pools
MONGO_MAX_POOL_SIZE=50
REDIS_POOL_SIZE=50
```

### Low Latency (< 1k req/s)

Edit `deployment/local/.env`:
```bash
# Disable buffer for immediate writes
LINK_BUFFER_ENABLED=false

# Optimize cache
CACHE_TTL=5m
CACHE_MAX_SIZE=10000
```

After changes, restart:
```bash
./janus restart api
```

## 🔐 Security Checklist

Before going to production:

- [ ] Change default API key in `.env`
- [ ] Change MongoDB password
- [ ] Use HTTPS in production (configure reverse proxy)
- [ ] Enable rate limiting (already configured)
- [ ] Configure firewall rules
- [ ] Set up monitoring and alerts
- [ ] Configure automated backups
- [ ] Review security headers in Nginx
- [ ] Enable Sentry error tracking (set `SENTRY_DSN`)
- [ ] Restrict MongoDB/Redis network access

## 📞 Support

- **Documentation:** README.md
- **Issues:** GitHub Issues
- **Tests:** TESTING.md
- **Deployment:** deployment/DEPLOYMENT.md

## 🎉 You're Ready!

Your Janus instance is now running. Start shortening URLs!

```bash
# Quick test
curl -X POST http://localhost:8083/api/v1/link \
  -H "Authorization: Bearer secret-api-key" \
  -H "Content-Type: application/json" \
  -d '{"link": "https://github.com"}'
```

Happy shortening! 🚀
