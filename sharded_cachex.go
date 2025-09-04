package cachex

import (
	"fmt"
	"hash/fnv"
	"time"
)

type ShardedCache[K comparable, V any] struct {
	shards  []*shard[K, V]
	janitor *janitor
}

// hashKey deterministically maps a comparable key to a shard index
func hashKey[K comparable](key K, numShards int) int {
	h := fnv.New32a()
	_, _ = h.Write([]byte(fmt.Sprintf("%v", key)))
	return int(h.Sum32()) % numShards
}

func NewSharded[K comparable, V any](numShards int, defaultTTL, cleanupInterval time.Duration) *ShardedCache[K, V] {
	sc := &ShardedCache[K, V]{
		shards: make([]*shard[K, V], numShards),
	}
	for i := range numShards {
		sc.shards[i] = newShard[K, V](defaultTTL)
	}

	if defaultTTL > 0 && cleanupInterval > 0 {
		j := newJanitor(cleanupInterval)
		sc.janitor = j
		j.run(sc)
	}
	return sc
}

func (sc *ShardedCache[K, V]) Set(key K, value V) {
	idx := hashKey(key, len(sc.shards))
	sc.shards[idx].set(key, value)
}

func (sc *ShardedCache[K, V]) Get(key K) (V, bool) {
	idx := hashKey(key, len(sc.shards))
	return sc.shards[idx].get(key)
}

func (sc *ShardedCache[K, V]) Delete(key K) {
	idx := hashKey(key, len(sc.shards))
	sc.shards[idx].delete(key)
}

func (sc *ShardedCache[K, V]) cleanup() {
	for _, s := range sc.shards {
		s.cleanup()
	}
}

func (sc *ShardedCache[K, V]) Close() {
	if sc.janitor != nil {
		sc.janitor.stopJanitor()
	}
}
