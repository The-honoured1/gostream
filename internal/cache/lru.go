package cache

import (
	"container/list"
	"sync"
)

// LRUCache implements a simple thread-safe size-limited cache.
type LRUCache struct {
	maxSize int64
	currSize int64
	ll      *list.List
	cache   map[string]*list.Element
	mu      sync.Mutex
}

type entry struct {
	key   string
	value []byte
}

func NewLRUCache(maxSize int64) *LRUCache {
	return &LRUCache{
		maxSize: maxSize,
		ll:      list.New(),
		cache:   make(map[string]*list.Element),
	}
}

func (c *LRUCache) Get(key string) ([]byte, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if ele, hit := c.cache[key]; hit {
		c.ll.MoveToFront(ele)
		return ele.Value.(*entry).value, true
	}
	return nil, false
}

func (c *LRUCache) Set(key string, value []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()

	valSize := int64(len(value))
	if valSize > c.maxSize {
		return // Too large for cache
	}

	if ele, ok := c.cache[key]; ok {
		c.ll.MoveToFront(ele)
		c.currSize -= int64(len(ele.Value.(*entry).value))
		ele.Value.(*entry).value = value
		c.currSize += valSize
	} else {
		ele := c.ll.PushFront(&entry{key, value})
		c.cache[key] = ele
		c.currSize += valSize
	}

	// Evict until size is within limits
	for c.currSize > c.maxSize {
		c.removeOldest()
	}
}

func (c *LRUCache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if ele, ok := c.cache[key]; ok {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeOldest() {
	ele := c.ll.Back()
	if ele != nil {
		c.removeElement(ele)
	}
}

func (c *LRUCache) removeElement(e *list.Element) {
	c.ll.Remove(e)
	kv := e.Value.(*entry)
	delete(c.cache, kv.key)
	c.currSize -= int64(len(kv.value))
}
