// Package cache provides a generic in-memory key-value cache with TTL expiry.
// Useful for short-lived data like flight search results, rate limit counters,
// or any computation that's expensive to repeat.
// Not suitable for persistent or distributed caching — use Redis for that.
package cache

import (
	"sync"
	"time"
)

type entry struct {
	value     any
	expiresAt time.Time
}

// MemoryCache is a thread-safe, TTL-based in-memory key-value store.
type MemoryCache struct {
	mu    sync.RWMutex
	items map[string]entry
}

// NewMemoryCache creates and returns a new MemoryCache.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{items: make(map[string]entry)}
}

// Set stores a value under key with the given TTL.
func (c *MemoryCache) Set(key string, value any, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items[key] = entry{value: value, expiresAt: time.Now().Add(ttl)}
}

// Get retrieves a value by key. Returns (nil, false) if missing or expired.
func (c *MemoryCache) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	e, ok := c.items[key]
	if !ok || time.Now().After(e.expiresAt) {
		return nil, false
	}
	return e.value, true
}

// Delete removes an entry by key.
func (c *MemoryCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
}

// Clear removes all expired entries and returns the count removed.
// Call this periodically from a background goroutine to prevent memory leaks.
func (c *MemoryCache) Clear() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	now := time.Now()
	removed := 0
	for k, e := range c.items {
		if now.After(e.expiresAt) {
			delete(c.items, k)
			removed++
		}
	}
	return removed
}

// Size returns the total number of entries (including expired but not yet evicted).
func (c *MemoryCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}
