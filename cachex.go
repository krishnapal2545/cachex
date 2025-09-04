package cachex

import (
	"sync"
	"time"
)

// Cache is a generic, in-memory, thread-safe cache with TTL support.
type Cache[K comparable, V any] struct {
	items   map[K]Item[V]
	mu      sync.RWMutex
	ttl     time.Duration
	janitor *janitor
}

// New creates a new cache instance with given default TTL and cleanup interval.
func New[K comparable, V any](defaultTTL, cleanupInterval time.Duration) *Cache[K, V] {
	c := &Cache[K, V]{
		items: make(map[K]Item[V]),
		ttl:   defaultTTL,
	}
	// only start janitor if TTL is > 0
	if defaultTTL > 0 && cleanupInterval > 0 {
		j := newJanitor(cleanupInterval)
		c.janitor = j
		j.run(c)
	}
	return c
}

// Set stores a key-value pair and resets its expiration based on default TTL.
func (c *Cache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	var exp int64
	if c.ttl > 0 {
		exp = time.Now().Add(c.ttl).UnixNano()
	}
	c.items[key] = Item[V]{Value: value, Expiration: exp}
}

// Get retrieves a value by key. If expired or missing, returns zero value.
func (c *Cache[K, V]) Get(key K) (V, bool) {
	var zero V

	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	if !found {
		return zero, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		c.Delete(key)
	}
	return item.Value, true
}

// Delete removes a key from the cache.
func (c *Cache[K, V]) Delete(key K) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Items returns a copy of all key-value pairs currently in the cache.
func (c *Cache[K, V]) Items() map[K]V {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[K]V, len(c.items))
	now := time.Now().UnixNano()

	for k, it := range c.items {
		if it.Expiration == 0 || it.Expiration > now {
			result[k] = it.Value
		}
	}
	return result
}

// must implement cleanupTarget
func (c *Cache[K, V]) cleanup() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for k, v := range c.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(c.items, k)
		}
	}
	c.mu.Unlock()
}

func (c *Cache[K, V]) Close() {
	if c.janitor != nil {
		c.janitor.stopJanitor()
	}
}
