# cachex

<img src="docs/squirrel_cache.png" alt="CacheX mascot" width="150" align="right">

🚀 A lightweight in-memory cache library for Go with **TTL (Time-To-Live)** and optional **sharding** for high concurrency.  

Supports:
- ✅ In-memory caching with optional expiration  
- ✅ Configurable default TTL (per cache)  
- ✅ Automatic cleanup with background janitor  
- ✅ Sharded cache (lock striping) for read/write heavy workloads  
- ✅ Pure Go (no dependencies, no CGO)  

---

## ✨ Features

- **Simple Cache with Expiry**
  ```go
    c := cachex.New[string, int](time.Minute, 10*time.Second)
    c.Set("foo", 42)
    v, ok := c.Get("foo")
  ```
- **No Expiry (always in memory)**
    ```go
    c := cachex.New[string, string](0, 0) // no TTL, no janitor
    c.Set("key", "value")

    ```
- **Sharded Cache**
    ```go
    sc := cachex.NewSharded[string, int](32, time.Minute, 30*time.Second)
    sc.Set("foo", 99)
    v, ok := sc.Get("foo")
    ```
- **Sharded Cache without Expiry**
    ```go
    sc := cachex.NewSharded[string, int](64, 0, 0) // heavy read/write cache
    ```

## 🔧 Installation
    go get github.com/krishnapal2545/cachex

## 📖 Usage Example

```go
    package main

    import (
        "fmt"
        "time"

        "github.com/krishnapal2545/cachex"
    )

    func main() {
        // Simple cache with TTL = 1 minute, cleanup every 30s
        c := cachex.New[string, int](time.Minute, 30*time.Second)
        defer c.Close()

        c.Set("answer", 42)

        if v, ok := c.Get("answer"); ok {
            fmt.Println("Got:", v)
        }

        // Sharded cache for high concurrency
        sc := cachex.NewSharded[string, int](32, 0, 0) // no expiry
        defer sc.Close()

        sc.Set("foo", 99)
        if v, ok := sc.Get("foo"); ok {
            fmt.Println("Sharded Got:", v)
        }
    }
```

## 🧪 Tests & Benchmarks

- *Run tests:*

    ```bash
    go test ./...
    ```


- *Run benchmarks:*

    ```bash 
    go test -bench=. -benchmem ./...
    ```


>  *Example benchmark results on 8-core machine (1M keys, 32 shards):*

- goos: windows
- goarch: amd64
- pkg: github.com/krishnapal2545/cachex
- cpu: Intel(R) Core(TM) Ultra 7 155U

    ```bash

    BenchmarkCacheConcurrentGet-8              16662207     69.26 ns/op
    BenchmarkCacheConcurrentReadWrite-8         8,668,077    203.3 ns/op
    BenchmarkShardedCacheConcurrentGet-8       14,101,520     80.77 ns/op
    BenchmarkShardedCacheConcurrentReadWrite-8 12,568,489     90.96 ns/op

    ```

##  ⚡ When to Use

- *Use Cache if:*

    - You need a simple in-memory cache

    - Workload is small (< 1M keys) or read-heavy

- *Use ShardedCache if:* 

    - You expect 1M–30M+ keys

    - Workload is read/write heavy with high concurrency

    - You want to reduce lock contention

## 🏗 Internals (Design)

- Item
Each cached value is wrapped in an Item struct that holds the value and expiration timestamp (0 means no expiry).

- Janitor
A background goroutine that periodically removes expired items.

    - Runs only if ttl > 0 && cleanupInterval > 0.

    - Not started when using a non-expiring cache.

- Cache
A simple single-shard cache protected by a sync.RWMutex.

- ShardedCache
Splits keys across N shards using FNV-1a hashing.

    -   Each shard has its own lock & map.

    - Greatly reduces lock contention for concurrent workloads.


## 📜 License

MIT License

---
