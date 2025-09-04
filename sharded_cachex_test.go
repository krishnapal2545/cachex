package cachex

import (
	"sync"
	"testing"
	"time"
)

func TestShardedCacheBasic(t *testing.T) {
	sc := NewSharded[string, int](16, time.Minute, time.Second*10)
	defer sc.Close()

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
	defer sc.Close()

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


// Benchmark: Concurrent Get on 1,000,000 items
func BenchmarkShardedCacheConcurrentGet(b *testing.B) {
	const N = 1_000_000

	sc := NewSharded[int, int](32, time.Minute, time.Second*30)
	defer sc.Close()

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
	defer sc.Close()

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
