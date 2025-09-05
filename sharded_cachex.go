package cachex

import (
	"fmt"
	"hash/fnv"
	"maps"
	"time"
)

type ShardedCache[K comparable, V any] struct {
	shards []*Cache[K, V]
}

// hashKey deterministically maps a comparable key to a shard index
func hashKey[K comparable](key K, numShards int) int {
	h := fnv.New32a()
	_, _ = fmt.Fprintf(h, "%v", key)
	return int(h.Sum32()) % numShards
}

func NewSharded[K comparable, V any](numShards int, defaultTTL, cleanupInterval time.Duration) *ShardedCache[K, V] {
	sc := &ShardedCache[K, V]{
		shards: make([]*Cache[K, V], numShards),
	}
	for i := range numShards {
		sc.shards[i] = New[K, V](defaultTTL, cleanupInterval)
	}
	return sc
}

func (sc *ShardedCache[K, V]) Set(key K, value V) {
	idx := hashKey(key, len(sc.shards))
	sc.shards[idx].Set(key, value)
}

func (sc *ShardedCache[K, V]) Get(key K) (V, bool) {
	idx := hashKey(key, len(sc.shards))
	return sc.shards[idx].Get(key)
}

func (sc *ShardedCache[K, V]) Delete(key K) {
	idx := hashKey(key, len(sc.shards))
	sc.shards[idx].Delete(key)
}

// Items returns a copy of all key-value pairs across all shards.
func (sc *ShardedCache[K, V]) Items() map[K]V {
	result := make(map[K]V)

	for _, shard := range sc.shards {
		items := shard.Items()
		maps.Copy(result, items)
	}

	return result
}

func (sc *ShardedCache[K, V]) Len() int {
	total := 0
	for _, shard := range sc.shards {
		total += shard.Len()
	}
	return total
}

func (sc *ShardedCache[K, V]) Range(f func(K, V) bool) {
	for _, shard := range sc.shards {
		shard.Range(f)
	}
}

func (sc *ShardedCache[K, V]) Clear() {
	for _, shard := range sc.shards {
		shard.Clear()
	}
}

func (sc *ShardedCache[K, V]) Close() {
	for _, shard := range sc.shards {
		shard.Close()
	}
}
