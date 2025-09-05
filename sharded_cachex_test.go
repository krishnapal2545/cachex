package cachex

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestShardedCacheBasic(t *testing.T) {
	sc := NewSharded[string, int](16, time.Minute, time.Second*10)

	sc.Set("foo", 42)

	if v, ok := sc.Get("foo"); !ok || v != 42 {
		t.Errorf("expected 42, got %v (ok=%v)", v, ok)
	}

	sc.Delete("foo")

	if _, ok := sc.Get("foo"); ok {
		t.Error("expected key foo to be deleted")
	}
}

func TestShardedCacheConcurrentReadWrite(t *testing.T) {
	const workers = 50
	const iterations = 5000

	sc := NewSharded[int, int](32, time.Minute, time.Second*10)

	var wg sync.WaitGroup
	wg.Add(workers)

	for w := 0; w < workers; w++ {
		go func(workerID int) {
			defer wg.Done()
			for i := range iterations {
				key := (workerID * iterations) + i
				sc.Set(key, i)

				if v, ok := sc.Get(key); ok && v != i {
					t.Errorf("expected %d, got %d", i, v)
				}

				if i%100 == 0 {
					sc.Delete(key)
				}
			}
		}(w)
	}

	wg.Wait()
}

func TestShardedCacheLen(t *testing.T) {
	sc := NewSharded[string, int](8, time.Minute, time.Minute)

	if sc.Len() != 0 {
		t.Fatalf("expected length 0, got %d", sc.Len())
	}

	// Distribute across shards
	for i := range 100 {
		sc.Set(string(rune('a'+(i%26))), i)
	}

	if sc.Len() != 26 { // keys repeat, so should stabilize at 26
		t.Fatalf("expected length 26, got %d", sc.Len())
	}

	sc.Delete("a")

	if sc.Len() != 25 {
		t.Fatalf("expected length 25 after delete, got %d", sc.Len())
	}
}

func TestShardedCacheMillionLen(t *testing.T) {
	sc := NewSharded[string, int](16, time.Minute, time.Minute)

	for i := 0; i < 1_000_000; i++ {
		// keys cycle through 26 unique values
		key := "key_" + strconv.Itoa(i) // unique key for every insert
		sc.Set(key, i)
	}

	start := time.Now()
	if sc.Len() != 1_000_000 {
		t.Fatalf("expected length 1,000,000, got %d", sc.Len())
	}
	elapsed := time.Since(start)

	fmt.Printf("Inserted 1,000,000 items into ShardedCache, final Len=%d, took=%s\n", sc.Len(), elapsed)
}

func TestNestedShardedCache(t *testing.T) {
	outer := NewSharded[string, any](8, 2*time.Second, time.Second)
	inner := NewSharded[string, string](4, 2*time.Second, time.Second)

	// Put something into the inner cache
	inner.Set("foo", "bar")

	// Insert inner sharded cache into the outer
	outer.Set("nested", inner)
	fmt.Println("Set nested sharded cache inside outer")

	// Delete manually
	outer.Delete("nested")

	// Reinsert for expiry test
	outer.Set("nested", inner)

	// Wait until janitor expiry runs
	time.Sleep(3 * time.Second)
}

func TestHybridNestedCache(t *testing.T) {
	// Outer: ShardedCache
	outer := NewSharded[string, any](8, 2*time.Second, time.Second)

	// Mid: Normal Cache
	mid := New[string, any](2*time.Second, time.Second)

	// Inner: ShardedCache
	inner := NewSharded[string, string](4, 2*time.Second, time.Second)

	// Put data into inner
	inner.Set("deepKey", "deepValue")

	// Nest inner -> mid
	mid.Set("innerSharded", inner)

	// Nest mid -> outer
	outer.Set("midCache", mid)

	fmt.Println("✅ Set hybrid nested cache: Sharded -> Cache -> Sharded")

	start := time.Now()
	// Manual delete
	outer.Delete("midCache")

	elapsed := time.Since(start)

	fmt.Printf("Deleted time taken : %v\n", elapsed)

	// Reinsert for expiry test
	outer.Set("midCache", mid)

	// Wait for janitor to trigger
	time.Sleep(3 * time.Second)
}

// Benchmark: Concurrent Get on 1,000,000 items
func BenchmarkShardedCacheConcurrentGet(b *testing.B) {
	const N = 1_000_000

	sc := NewSharded[int, int](32, time.Minute, time.Second*30)

	// preload data
	for i := range N {
		sc.Set(i, i*2)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			key := time.Now().Nanosecond() % N
			sc.Get(key)
		}
	})
}

// Benchmark: Concurrent Read + Write + Delete
func BenchmarkShardedCacheConcurrentReadWrite(b *testing.B) {
	const N = 1_000_000
	sc := NewSharded[int, int](32, time.Minute, time.Second*30)

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := i % N
			sc.Set(key, i)
			sc.Get(key)
			if i%50 == 0 {
				sc.Delete(key)
			}
			i++
		}
	})
}
