package service

import (
	"sync"
	"time"
)

type LinkCacheConfig struct {
	TTL             time.Duration // How long entries should live (e.g., 5*time.Minute)
	CleanupInterval time.Duration // How often to cleanup expired buckets (e.g., 1*time.Minute)
	MaxSize         int           // Maximum number of entries (0 for unlimited)
}

// LinkCache provides in-memory caching with time-bucketed expiration
type LinkCache struct {
	cache           map[int64]map[string]string // map[expiryMinute]map[shortHash]originalURL
	mu              sync.RWMutex
	ttl             time.Duration
	cleanupInterval time.Duration
	maxSize         int // Maximum number of entries (0 for unlimited)
}

var (
	cacheInstance *LinkCache
	cacheOnce     sync.Once
)

// InitLinkCache initializes the global link cache instance
func InitLinkCache(cfg LinkCacheConfig) *LinkCache {
	cacheOnce.Do(func() {
		cacheInstance = &LinkCache{
			cache:           make(map[int64]map[string]string),
			ttl:             cfg.TTL,
			cleanupInterval: cfg.CleanupInterval,
			maxSize:         cfg.MaxSize,
		}
		// Start cleanup goroutine to remove expired time buckets
		go cacheInstance.cleanupExpired()
	})
	return cacheInstance
}

// GetLinkCache returns the global link cache instance
func GetLinkCache() *LinkCache {
	return cacheInstance
}

// Get retrieves a link from cache
func (lc *LinkCache) Get(shortLinkHash string) (string, bool) {
	lc.mu.RLock()
	defer lc.mu.RUnlock()

	// Search through all time buckets
	for _, bucket := range lc.cache {
		if originalURL, exists := bucket[shortLinkHash]; exists {
			return originalURL, true
		}
	}

	return "", false
}

// Set stores a link in cache
func (lc *LinkCache) Set(shortLinkHash, originalURL string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	if lc.maxSize > 0 && lc.size() >= lc.maxSize {
		lc.removeOldest()
	}

	expiryTime := time.Now().Add(lc.ttl)
	expiryMinute := expiryTime.Unix() / 60

	// Initialize bucket if it doesn't exist
	if lc.cache[expiryMinute] == nil {
		lc.cache[expiryMinute] = make(map[string]string)
	}

	lc.cache[expiryMinute][shortLinkHash] = originalURL
}

// Delete removes a link from cache
func (lc *LinkCache) Delete(shortLinkHash string) {
	lc.mu.Lock()
	defer lc.mu.Unlock()

	// Remove from all buckets (usually just one)
	for _, bucket := range lc.cache {
		delete(bucket, shortLinkHash)
	}
}

// cleanupExpired removes expired time buckets periodically
func (lc *LinkCache) cleanupExpired() {
	ticker := time.NewTicker(lc.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		lc.mu.Lock()
		now := time.Now()
		currentMinute := now.Unix() / 60

		// Remove all buckets that have expired
		for expiryMinute := range lc.cache {
			if expiryMinute <= currentMinute {
				delete(lc.cache, expiryMinute)
			}
		}
		lc.mu.Unlock()
	}
}

// removeOldest removes the oldest expiry bucket to make room for new entries
func (lc *LinkCache) removeOldest() {
	if len(lc.cache) == 0 {
		return
	}

	// Find the oldest bucket
	var oldestMinute int64 = -1
	for minute := range lc.cache {
		if oldestMinute == -1 || minute < oldestMinute {
			oldestMinute = minute
		}
	}

	// Remove the oldest bucket
	if oldestMinute != -1 {
		delete(lc.cache, oldestMinute)
	}
}

// size returns the current number of cached entries (must be called with lock held)
func (lc *LinkCache) size() int {
	count := 0
	for _, bucket := range lc.cache {
		count += len(bucket)
	}
	return count
}

// Size returns the current number of cached entries
func (lc *LinkCache) Size() int {
	lc.mu.RLock()
	defer lc.mu.RUnlock()
	return lc.size()
}

// Clear removes all entries from cache
func (lc *LinkCache) Clear() {
	lc.mu.Lock()
	defer lc.mu.Unlock()
	lc.cache = make(map[int64]map[string]string)
}
