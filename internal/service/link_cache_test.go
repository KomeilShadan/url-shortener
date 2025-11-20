package service

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestInitLinkCache(t *testing.T) {
	// Reset singleton for testing
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cfg := LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         1000,
	}

	cache := InitLinkCache(cfg)

	if cache == nil {
		t.Fatal("Expected cache to be initialized")
	}

	if cache.ttl != cfg.TTL {
		t.Errorf("Expected TTL %v, got %v", cfg.TTL, cache.ttl)
	}

	if cache.maxSize != cfg.MaxSize {
		t.Errorf("Expected MaxSize %d, got %d", cfg.MaxSize, cache.maxSize)
	}

	// Test singleton pattern
	cache2 := InitLinkCache(LinkCacheConfig{TTL: 10 * time.Minute})
	if cache2 != cache {
		t.Error("Expected same cache instance (singleton)")
	}
}

func TestLinkCache_SetAndGet(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	// Set a value
	cache.Set("abc123", "https://example.com")

	// Get the value
	url, found := cache.Get("abc123")

	if !found {
		t.Error("Expected to find cached value")
	}

	if url != "https://example.com" {
		t.Errorf("Expected 'https://example.com', got '%s'", url)
	}
}

func TestLinkCache_GetNonExistent(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	url, found := cache.Get("nonexistent")

	if found {
		t.Error("Expected not to find non-existent key")
	}

	if url != "" {
		t.Errorf("Expected empty string, got '%s'", url)
	}
}

func TestLinkCache_Delete(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	// Set and verify
	cache.Set("abc123", "https://example.com")
	_, found := cache.Get("abc123")
	if !found {
		t.Fatal("Expected to find cached value before delete")
	}

	// Delete
	cache.Delete("abc123")

	// Verify deleted
	_, found = cache.Get("abc123")
	if found {
		t.Error("Expected not to find deleted value")
	}
}

func TestLinkCache_Size(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	if cache.Size() != 0 {
		t.Errorf("Expected size 0, got %d", cache.Size())
	}

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Errorf("Expected size 3, got %d", cache.Size())
	}
}

func TestLinkCache_Clear(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	// Add some entries
	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Set("key3", "value3")

	if cache.Size() != 3 {
		t.Fatalf("Expected size 3 before clear, got %d", cache.Size())
	}

	// Clear
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}

	// Verify entries are gone
	_, found := cache.Get("key1")
	if found {
		t.Error("Expected not to find cleared entry")
	}
}

func TestLinkCache_MaxSize(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	maxSize := 5
	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         maxSize,
	})

	// Add more entries than maxSize
	for i := 0; i < maxSize+3; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	// Size should not exceed maxSize
	if cache.Size() > maxSize {
		t.Errorf("Expected size <= %d, got %d", maxSize, cache.Size())
	}
}

func TestLinkCache_Expiration(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             100 * time.Millisecond,
		CleanupInterval: 50 * time.Millisecond,
		MaxSize:         100,
	})

	// Set a value
	cache.Set("abc123", "https://example.com")

	// Verify it exists
	_, found := cache.Get("abc123")
	if !found {
		t.Fatal("Expected to find cached value immediately")
	}

	// Wait for expiration + cleanup
	time.Sleep(200 * time.Millisecond)

	// Verify it's gone
	_, found = cache.Get("abc123")
	if found {
		t.Error("Expected cached value to be expired and cleaned up")
	}
}

func TestLinkCache_ConcurrentAccess(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         1000,
	})

	const numGoroutines = 50
	const opsPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines * 3) // Set, Get, Delete operations

	// Concurrent Set operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				value := fmt.Sprintf("value_%d_%d", id, j)
				cache.Set(key, value)
			}
		}(i)
	}

	// Concurrent Get operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Get(key)
			}
		}(i)
	}

	// Concurrent Delete operations
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key_%d_%d", id, j)
				cache.Delete(key)
			}
		}(i)
	}

	wg.Wait()

	// No assertions needed - just verify no race conditions or panics
}

func TestLinkCache_UpdateExisting(t *testing.T) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         100,
	})

	// Set initial value
	cache.Set("abc123", "https://example.com")

	// Update with new value
	cache.Set("abc123", "https://updated.com")

	// Get and verify updated value
	url, found := cache.Get("abc123")

	if !found {
		t.Fatal("Expected to find cached value")
	}

	if url != "https://updated.com" {
		t.Errorf("Expected 'https://updated.com', got '%s'", url)
	}
}

func BenchmarkLinkCache_Set(b *testing.B) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         10000,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}
}

func BenchmarkLinkCache_Get(b *testing.B) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         10000,
	})

	// Pre-populate cache
	for i := 0; i < 1000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), fmt.Sprintf("value%d", i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key%d", i%1000))
	}
}

func BenchmarkLinkCache_ConcurrentSetGet(b *testing.B) {
	// Reset singleton
	cacheInstance = nil
	cacheOnce = sync.Once{}

	cache := InitLinkCache(LinkCacheConfig{
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		MaxSize:         10000,
	})

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := fmt.Sprintf("key%d", i%1000)
			if i%2 == 0 {
				cache.Set(key, fmt.Sprintf("value%d", i))
			} else {
				cache.Get(key)
			}
			i++
		}
	})
}
