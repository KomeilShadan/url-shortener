package service

import (
	"context"
	"fmt"
	"janus/internal/repository"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	testMongoURI = "mongodb://localhost:27017"
	testDBName   = "janus_test"
)

// setupTestRepo creates a test MongoDB connection and repository
func setupTestRepo(t *testing.T) (*mongo.Client, repository.LinkRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(testMongoURI))
	if err != nil {
		t.Skipf("Skipping test: MongoDB not available: %v", err)
		return nil, nil
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		t.Skipf("Skipping test: MongoDB not reachable: %v", err)
		return nil, nil
	}

	// Clean up test database
	if err := client.Database(testDBName).Drop(ctx); err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
	}

	// Create repository
	repo := repository.NewMongoLinkRepository(client, testDBName)

	return client, repo
}

// teardownTestRepo cleans up test MongoDB connection
func teardownTestRepo(t *testing.T, client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Drop test database
	if err := client.Database(testDBName).Drop(ctx); err != nil {
		t.Logf("Warning: Failed to drop test database: %v", err)
	}

	// Disconnect
	if err := client.Disconnect(ctx); err != nil {
		t.Logf("Warning: Failed to disconnect from MongoDB: %v", err)
	}
}

func TestNewLinkBuffer(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 100, 5)

	if buffer == nil {
		t.Fatal("Expected buffer to be initialized")
	}

	if buffer.workerPool != 5 {
		t.Errorf("Expected 5 workers, got %d", buffer.workerPool)
	}

	if cap(buffer.buffer) != 100 {
		t.Errorf("Expected buffer size 100, got %d", cap(buffer.buffer))
	}

	if !buffer.enabled {
		t.Error("Expected buffer to be enabled")
	}

	buffer.Shutdown(2 * time.Second)
}

func TestLinkBuffer_QueueLink(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 10, 2)
	defer buffer.Shutdown(2 * time.Second)

	// Queue a link
	err := buffer.QueueLink("https://example.com", "abc123")
	if err != nil {
		t.Fatalf("Failed to queue link: %v", err)
	}

	// Give workers time to process
	time.Sleep(100 * time.Millisecond)

	// Verify link was inserted using repository
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	link, err := repo.FindByHash(ctx, "abc123")
	if err != nil {
		t.Fatalf("Failed to find inserted link: %v", err)
	}

	if link.Link != "https://example.com" {
		t.Errorf("Expected link 'https://example.com', got '%v'", link.Link)
	}
}

func TestLinkBuffer_QueueMultipleLinks(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 100, 5)
	defer buffer.Shutdown(2 * time.Second)

	// Queue multiple links
	links := []struct {
		url  string
		hash string
	}{
		{"https://example1.com", "hash1"},
		{"https://example2.com", "hash2"},
		{"https://example3.com", "hash3"},
		{"https://example4.com", "hash4"},
		{"https://example5.com", "hash5"},
	}

	for _, link := range links {
		err := buffer.QueueLink(link.url, link.hash)
		if err != nil {
			t.Fatalf("Failed to queue link %s: %v", link.hash, err)
		}
	}

	// Give workers time to process all jobs
	time.Sleep(500 * time.Millisecond)

	// Verify all links were inserted
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := client.Database(testDBName).Collection("links").
		CountDocuments(ctx, bson.M{})

	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	if count != int64(len(links)) {
		t.Errorf("Expected %d documents, got %d", len(links), count)
	}
}

func TestLinkBuffer_BufferFull(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	// Create buffer with very small size and no workers (to fill it up)
	buffer := NewLinkBuffer(repo, 2, 0)
	defer buffer.Shutdown(2 * time.Second)

	// Fill the buffer
	buffer.QueueLink("https://example1.com", "hash1")
	buffer.QueueLink("https://example2.com", "hash2")

	// This should trigger synchronous processing
	err := buffer.QueueLink("https://example3.com", "hash3")
	if err != nil {
		t.Fatalf("Failed to queue link when buffer full: %v", err)
	}

	// Give time for synchronous processing
	time.Sleep(200 * time.Millisecond)

	// Verify the synchronously processed link was inserted using repository
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	link, err := repo.FindByHash(ctx, "hash3")
	if err != nil {
		t.Fatalf("Failed to find synchronously inserted link: %v", err)
	}

	if link.ShortLinkHash != "hash3" {
		t.Errorf("Expected hash 'hash3', got '%v'", link.ShortLinkHash)
	}
}

func TestLinkBuffer_Upsert(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 10, 2)
	defer buffer.Shutdown(2 * time.Second)

	// Insert initial link
	err := buffer.QueueLink("https://example.com", "abc123")
	if err != nil {
		return
	}
	time.Sleep(100 * time.Millisecond)

	// Update with same link (should upsert)
	err = buffer.QueueLink("https://example.com", "abc123")
	if err != nil {
		return
	}
	time.Sleep(100 * time.Millisecond)

	// Verify only one document exists
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	count, err := client.Database(testDBName).Collection("links").
		CountDocuments(ctx, bson.M{"short_link_hash": "abc123"})

	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	if count != 1 {
		t.Errorf("Expected 1 document after upsert, got %d", count)
	}
}

func TestLinkBuffer_Shutdown(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 100, 5)

	// Queue some links
	for i := 0; i < 10; i++ {
		buffer.QueueLink("https://example.com", "hash")
	}

	// Shutdown with reasonable timeout
	buffer.Shutdown(3 * time.Second)

	// Verify workers stopped (buffer should be closed)
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when sending to closed channel")
		}
	}()

	buffer.buffer <- LinkInsertJob{}
}

func TestLinkBuffer_Stats(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 100, 5)
	defer buffer.Shutdown(2 * time.Second)

	stats := buffer.Stats()

	if stats["buffer_size"] != 100 {
		t.Errorf("Expected buffer_size 100, got %v", stats["buffer_size"])
	}

	if stats["worker_count"] != 5 {
		t.Errorf("Expected worker_count 5, got %v", stats["worker_count"])
	}

	if _, ok := stats["pending_jobs"]; !ok {
		t.Error("Expected pending_jobs in stats")
	}

	if _, ok := stats["buffer_usage_pct"]; !ok {
		t.Error("Expected buffer_usage_pct in stats")
	}
}

func TestLinkBuffer_ConcurrentQueueing(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	buffer := NewLinkBuffer(repo, 1000, 10)
	defer buffer.Shutdown(5 * time.Second)

	// Concurrent queueing
	const numGoroutines = 50
	const linksPerGoroutine = 10

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < linksPerGoroutine; j++ {
				hash := fmt.Sprintf("hash_%d_%d", id, j)
				url := fmt.Sprintf("https://example%d-%d.com", id, j)
				buffer.QueueLink(url, hash)
			}
		}(i)
	}

	wg.Wait()

	// Give workers time to process
	time.Sleep(2 * time.Second)

	// Verify all links were inserted
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := client.Database(testDBName).Collection("links").
		CountDocuments(ctx, bson.M{})

	if err != nil {
		t.Fatalf("Failed to count documents: %v", err)
	}

	expected := int64(numGoroutines * linksPerGoroutine)
	if count != expected {
		t.Errorf("Expected %d documents, got %d", expected, count)
	}
}

func TestNoOpBuffer(t *testing.T) {
	buffer := &NoOpBuffer{}

	// Test QueueLink (should be no-op)
	err := buffer.QueueLink("https://example.com", "abc123")
	if err != nil {
		t.Errorf("Expected no error from NoOpBuffer, got: %v", err)
	}

	// Test Shutdown (should be no-op)
	buffer.Shutdown(1 * time.Second)

	// Test Stats
	stats := buffer.Stats()
	if stats["enabled"] != false {
		t.Error("Expected NoOpBuffer to be disabled")
	}
	if stats["type"] != "noop" {
		t.Errorf("Expected type 'noop', got: %v", stats["type"])
	}
}

func TestLinkBuffer_Interface(t *testing.T) {
	client, repo := setupTestRepo(t)
	if client == nil {
		return
	}
	defer teardownTestRepo(t, client)

	// Test that LinkBuffer implements LinkBufferInterface
	var _ LinkBufferInterface = NewLinkBuffer(repo, 10, 2)

	// Test that NoOpBuffer implements LinkBufferInterface
	var _ LinkBufferInterface = &NoOpBuffer{}
}
