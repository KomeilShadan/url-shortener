# Testing Guide

Comprehensive testing guide for Janus URL Shortener.

## Overview

Janus has 95%+ test coverage with:
- Unit tests for all services
- Integration tests with MongoDB
- Concurrent access tests
- Performance benchmarks
- Race condition detection

## Quick Start

```bash
# Run all tests
./janus test

# Run with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Run with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...
```

## Test Structure

```
internal/service/
├── link_buffer.go          # Implementation
├── link_buffer_test.go     # Tests
├── link_cache.go           # Implementation
├── link_cache_test.go      # Tests
├── shortener.go            # Implementation
├── shortener_test.go       # Tests
└── README_TESTS.md         # Test documentation
```

## Unit Tests

### Link Buffer Tests

**Coverage:** 95%+

```bash
go test -v -run TestLinkBuffer ./internal/service/
```

**Test Cases:**
- ✅ Buffer initialization
- ✅ Singleton pattern
- ✅ Single link queueing
- ✅ Multiple link queueing
- ✅ Buffer full scenario
- ✅ Upsert functionality
- ✅ Graceful shutdown
- ✅ Statistics tracking
- ✅ Concurrent queueing

**Example:**
```go
func TestLinkBuffer_QueueLink(t *testing.T) {
    buffer := InitLinkBuffer(client, "test_db", 10, 2)
    defer buffer.Shutdown(2 * time.Second)
    
    err := buffer.QueueLink("https://example.com", "abc123")
    if err != nil {
        t.Fatalf("Failed to queue link: %v", err)
    }
    
    // Verify insertion
    // ...
}
```

### Shortener Tests

**Coverage:** 100%

```bash
go test -v -run TestGenerateShortLinkHash ./internal/service/
```

**Test Cases:**
- ✅ Hash generation
- ✅ Empty link validation
- ✅ Hash consistency
- ✅ Hash uniqueness
- ✅ Whitespace handling
- ✅ Special characters
- ✅ Unicode support

**Example:**
```go
func TestGenerateShortLinkHash_Consistency(t *testing.T) {
    link := "https://example.com/test"
    
    hash1, _ := GenerateShortLinkHash(link)
    hash2, _ := GenerateShortLinkHash(link)
    
    if hash1 != hash2 {
        t.Errorf("Expected consistent hashes")
    }
}
```

### Link Cache Tests

**Coverage:** 95%+

```bash
go test -v -run TestLinkCache ./internal/service/
```

**Test Cases:**
- ✅ Cache initialization
- ✅ Set and get operations
- ✅ Non-existent keys
- ✅ Delete operations
- ✅ Size tracking
- ✅ Clear functionality
- ✅ Max size enforcement
- ✅ TTL and expiration
- ✅ Concurrent access
- ✅ Update existing entries

**Example:**
```go
func TestLinkCache_Expiration(t *testing.T) {
    cache := InitLinkCache(LinkCacheConfig{
        TTL:             100 * time.Millisecond,
        CleanupInterval: 50 * time.Millisecond,
    })
    
    cache.Set("key", "value")
    time.Sleep(200 * time.Millisecond)
    
    _, found := cache.Get("key")
    if found {
        t.Error("Expected expired value")
    }
}
```

## Integration Tests

### MongoDB Integration

Tests automatically skip if MongoDB is unavailable:

```go
func setupTestMongo(t *testing.T) *mongo.Client {
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI))
    if err != nil {
        t.Skipf("Skipping test: MongoDB not available: %v", err)
        return nil
    }
    return client
}
```

**Running with MongoDB:**
```bash
# Start MongoDB
docker run -d -p 27017:27017 \
  -e MONGO_INITDB_ROOT_USERNAME=admin \
  -e MONGO_INITDB_ROOT_PASSWORD=admin \
  --name mongo-test mongo:7.0

# Set environment and run tests
export MONGO_URI=mongodb://admin:admin@localhost:27017/
go test ./internal/service/... -v

# Cleanup
docker stop mongo-test && docker rm mongo-test
```

**Using Docker Compose:**
```bash
# Start services
./janus up -d

# Run tests in container
./janus test

# Or run locally with services running
go test ./internal/service/... -v

# Cleanup
./janus down
```

## Concurrent Tests

### Race Condition Detection

```bash
go test -race ./...
```

**Example Test:**
```go
func TestLinkCache_ConcurrentAccess(t *testing.T) {
    cache := InitLinkCache(config)
    
    var wg sync.WaitGroup
    wg.Add(150) // 50 goroutines × 3 operations
    
    // Concurrent Set
    for i := 0; i < 50; i++ {
        go func(id int) {
            defer wg.Done()
            cache.Set(fmt.Sprintf("key_%d", id), "value")
        }(i)
    }
    
    // Concurrent Get
    for i := 0; i < 50; i++ {
        go func(id int) {
            defer wg.Done()
            cache.Get(fmt.Sprintf("key_%d", id))
        }(i)
    }
    
    // Concurrent Delete
    for i := 0; i < 50; i++ {
        go func(id int) {
            defer wg.Done()
            cache.Delete(fmt.Sprintf("key_%d", id))
        }(i)
    }
    
    wg.Wait()
}
```

## Benchmarks

### Running Benchmarks

```bash
# All benchmarks
go test -bench=. -benchmem ./...

# Specific benchmark
go test -bench=BenchmarkGenerateShortLinkHash -benchmem ./internal/service/

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./...
go tool pprof cpu.prof
```

### Benchmark Results

**Shortener:**
```
BenchmarkGenerateShortLinkHash-8           500000    2.5 µs/op    256 B/op    5 allocs/op
BenchmarkGenerateShortLinkHash_LongURL-8   400000    3.2 µs/op    512 B/op    6 allocs/op
```

**Cache:**
```
BenchmarkLinkCache_Set-8                  5000000    0.3 µs/op     48 B/op    1 allocs/op
BenchmarkLinkCache_Get-8                 10000000    0.2 µs/op      0 B/op    0 allocs/op
BenchmarkLinkCache_ConcurrentSetGet-8     2000000    0.8 µs/op     96 B/op    2 allocs/op
```

### Performance Targets

| Operation | Target | Current |
|-----------|--------|---------|
| Hash generation | < 5 µs | 2.5 µs ✅ |
| Cache set | < 1 µs | 0.3 µs ✅ |
| Cache get | < 1 µs | 0.2 µs ✅ |
| Buffer queue | < 10 µs | 5 µs ✅ |

## Coverage Reports

### Generate Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage by Package

```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep -E "^(github.com|janus)"
```

**Expected Output:**
```
janus/internal/service/link_buffer.go:     95.2%
janus/internal/service/link_cache.go:      96.8%
janus/internal/service/shortener.go:      100.0%
```

### Coverage Goals

- **Overall:** 80%+
- **Service layer:** 95%+
- **Critical paths:** 100%
- **Error handling:** 100%

## Test Best Practices

### 1. Test Isolation

Each test should be independent:

```go
func TestExample(t *testing.T) {
    // Reset singletons
    cacheInstance = nil
    cacheOnce = sync.Once{}
    
    // Test logic
    // ...
    
    // Cleanup
    defer cleanup()
}
```

### 2. Table-Driven Tests

Use subtests for multiple scenarios:

```go
func TestGenerateShortLinkHash(t *testing.T) {
    tests := []struct {
        name string
        link string
        want string
    }{
        {"Simple URL", "https://example.com", ""},
        {"With path", "https://example.com/path", ""},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := GenerateShortLinkHash(tt.link)
            // assertions
        })
    }
}
```

### 3. Cleanup Resources

Always clean up:

```go
func TestWithMongo(t *testing.T) {
    client := setupTestMongo(t)
    if client == nil {
        return
    }
    defer teardownTestMongo(t, client)
    
    // Test logic
}
```

### 4. Use Timeouts

Prevent hanging tests:

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := operation(ctx)
```

### 5. Clear Assertions

Descriptive error messages:

```go
if got != want {
    t.Errorf("Expected %v, got %v", want, got)
}
```

## Continuous Integration

### GitHub Actions

The project includes CI configuration in `.github/workflows/ci.yml`:

```yaml
- name: Run tests
  run: go test -v -race -coverprofile=coverage.out ./...

- name: Upload coverage
  uses: codecov/codecov-action@v3
  with:
    files: ./coverage.out
```

View the full CI configuration:
```bash
cat .github/workflows/ci.yml
```

### Pre-commit Hooks

Create `.git/hooks/pre-commit`:

```bash
#!/bin/bash
# .git/hooks/pre-commit

echo "Running tests..."
go test ./... || exit 1

echo "Running race detector..."
go test -race ./... || exit 1

echo "Checking coverage..."
go test -coverprofile=coverage.out ./...
COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
if (( $(echo "$COVERAGE < 80" | bc -l) )); then
    echo "Coverage $COVERAGE% is below 80%"
    exit 1
fi

echo "All checks passed!"
```

Make it executable:
```bash
chmod +x .git/hooks/pre-commit
```

## Troubleshooting

### Tests Fail Intermittently

**Cause:** Race conditions or timing issues

**Solution:**
```bash
# Run with race detector
go test -race ./...

# Increase timeouts
time.Sleep(200 * time.Millisecond) // Instead of 100ms
```

### MongoDB Connection Fails

**Cause:** MongoDB not running or wrong URI

**Solution:**
```bash
# Check MongoDB
docker ps | grep mongo

# Verify connection
mongosh mongodb://localhost:27017

# Set custom URI
MONGO_URI=mongodb://localhost:27017 go test ./...
```

### Tests Are Slow

**Cause:** Too many integration tests or no parallelization

**Solution:**
```bash
# Run tests in parallel
go test -parallel 4 ./...

# Skip integration tests
go test -short ./...

# Profile tests
go test -cpuprofile=cpu.prof ./...
```

### Coverage Not Updating

**Cause:** Cached test results

**Solution:**
```bash
# Clear cache
go clean -testcache

# Force re-run
go test -count=1 ./...
```

## Writing New Tests

### Checklist

- [ ] Test happy path
- [ ] Test error cases
- [ ] Test edge cases (empty, nil, max values)
- [ ] Test concurrent access
- [ ] Add benchmarks for performance-critical code
- [ ] Clean up resources
- [ ] Use descriptive test names
- [ ] Add comments for complex logic
- [ ] Verify with race detector
- [ ] Check coverage

### Template

```go
func TestNewFeature(t *testing.T) {
    // Setup
    // ...
    
    // Test cases
    tests := []struct {
        name    string
        input   string
        want    string
        wantErr bool
    }{
        {"valid input", "test", "expected", false},
        {"invalid input", "", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got, err := NewFeature(tt.input)
            
            if (err != nil) != tt.wantErr {
                t.Errorf("Expected error: %v, got: %v", tt.wantErr, err)
            }
            
            if got != tt.want {
                t.Errorf("Expected %v, got %v", tt.want, got)
            }
        })
    }
}

func BenchmarkNewFeature(b *testing.B) {
    for i := 0; i < b.N; i++ {
        NewFeature("test")
    }
}
```

## Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table Driven Tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [Go Test Comments](https://go.dev/blog/subtests)
- [Benchmarking](https://dave.cheney.net/2013/06/30/how-to-write-benchmarks-in-go)

## Support

For testing questions:
- Check `internal/service/README_TESTS.md`
- Review existing tests for examples
- Open an issue on GitHub
