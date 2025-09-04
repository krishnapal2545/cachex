package cachex

import (
	"sync"
	"time"
)

type shard[K comparable, V any] struct {
	items map[K]Item[V]
	mu    sync.RWMutex
	ttl   time.Duration
}

func newShard[K comparable, V any](ttl time.Duration) *shard[K, V] {
	return &shard[K, V]{
		items: make(map[K]Item[V]),
		ttl:   ttl,
	}
}

func (s *shard[K, V]) set(key K, value V) {
	s.mu.Lock()
	defer s.mu.Unlock()

	exp := time.Now().Add(s.ttl).UnixNano()
	s.items[key] = Item[V]{
		Value:      value,
		Expiration: exp,
	}
}

func (s *shard[K, V]) get(key K) (V, bool) {
	var zero V
	s.mu.RLock()
	item, ok := s.items[key]
	s.mu.RUnlock()
	if !ok {
		return zero, false
	}
	if item.Expiration > 0 && time.Now().UnixNano() > item.Expiration {
		s.delete(key)
	}
	return item.Value, true
}

func (s *shard[K, V]) delete(key K) {
	s.mu.Lock()
	delete(s.items, key)
	s.mu.Unlock()
}

func (s *shard[K, V]) cleanup() {
	now := time.Now().UnixNano()
	s.mu.Lock()
	for k, v := range s.items {
		if v.Expiration > 0 && now > v.Expiration {
			delete(s.items, k)
		}
	}
	s.mu.Unlock()
}
