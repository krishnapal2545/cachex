package cachex

import "time"

// Item represents a cached value with an expiration time.
type Item[V any] struct {
	Value      V
	Expiration int64 // 0 means never expires
}

type janitor struct {
	interval time.Duration
	stop     chan struct{}
}

// cleanupTarget is the interface that anything using janitor must implement.
type cleanupTarget interface {
	cleanup()
}

type Closable interface {
	Close()
}
