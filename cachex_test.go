package cachex

import (
	"strconv"
	"sync"
	"testing"
	"time"
)

// ---------- Unit Tests ----------

func TestSetGet(t *testing.T) {
	c := New[string, int](time.Minute, time.Second*10)
	defer c.Close()

	c.Set("foo", 42)
	v, ok := c.Get("foo")
	if !ok || v != 42 {
		t.Errorf("expected 42, got %v (ok=%v)", v, ok)
	}
}

func TestDelete(t *testing.T) {
	c := New[string, string](time.Minute, time.Second*10)
	defer c.Close()

	c.Set("key", "value")
	c.Delete("key")
	if _, ok := c.Get("key"); ok {
		t.Errorf("expected key to be deleted")
	}
}

func TestExpiration(t *testing.T) {
	c := New[string, string](time.Millisecond*50, time.Millisecond*10)
	defer c.Close()

	c.Set("temp", "data")
	time.Sleep(100 * time.Millisecond)

	if _, ok := c.Get("temp"); ok {
		t.Errorf("expected key to expire")
	}
}

// ---------- Benchmark ----------

// Benchmark with 1,000,000 items and concurrent Get
func BenchmarkCacheConcurrentGet(b *testing.B) {
	const N = 1_000_000

	c := New[int, int](time.Minute, time.Second*30)
	defer c.Close()

	// preload data
	for i := range N {
		c.Set(i, i*2)
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Random access pattern using modulo
			key := time.Now().Nanosecond() % N
			if _, ok := c.Get(key); !ok {
				b.Errorf("key %d missing", key)
			}
		}
	})
}

// Benchmark: Concurrent Read + Write + Delete
func BenchmarkCacheConcurrentReadWrite(b *testing.B) {
	const N = 1_000_000
	c := New[int, int](time.Minute, time.Second*30)
	defer c.Close()

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			key := i % N
			c.Set(key, i)
			c.Get(key)
			if i%50 == 0 { // occasionally delete
				c.Delete(key)
			}
			i++
		}
	})
}

// ---------- Stress Test ----------

func TestConcurrentAccess(t *testing.T) {
	const N = 1000
	c := New[string, int](time.Minute, time.Second*10)
	defer c.Close()

	var wg sync.WaitGroup
	for i := 0; i < N; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := "k" + strconv.Itoa(i)
			c.Set(key, i)
			if v, ok := c.Get(key); !ok || v != i {
				t.Errorf("unexpected value for %s: %v (ok=%v)", key, v, ok)
			}
		}(i)
	}
	wg.Wait()
}
