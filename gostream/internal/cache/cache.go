package cache

import (
	"sync"
	"time"
)

// CacheEntry represents a cached item with an expiration time.
type CacheEntry struct {
	Data      []byte
	ExpiresAt time.Time
}

// MemoryCache is a simple thread-safe in-memory cache.
type MemoryCache struct {
	sync.RWMutex
	items map[string]CacheEntry
	ttl   time.Duration
}

// NewMemoryCache creates a new MemoryCache with a default TTL.
func NewMemoryCache(ttl time.Duration) *MemoryCache {
	return &MemoryCache{
		items: make(map[string]CacheEntry),
		ttl:   ttl,
	}
}

// Set adds an item to the cache.
func (c *MemoryCache) Set(key string, data []byte) {
	c.Lock()
	defer c.Unlock()

	c.items[key] = CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

// Get retrieves an item from the cache if it exists and hasn't expired.
func (c *MemoryCache) Get(key string) ([]byte, bool) {
	c.RLock()
	defer c.RUnlock()

	item, exists := c.items[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(item.ExpiresAt) {
		return nil, false
	}

	return item.Data, true
}

// Delete removes an item from the cache.
func (c *MemoryCache) Delete(key string) {
	c.Lock()
	defer c.Unlock()
	delete(c.items, key)
}

// Cleanup removes expired items from the cache.
func (c *MemoryCache) Cleanup() {
	c.Lock()
	defer c.Unlock()

	now := time.Now()
	for k, v := range c.items {
		if now.After(v.ExpiresAt) {
			delete(c.items, k)
		}
	}
}
