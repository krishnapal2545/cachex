package cachex

import (
	"time"
)

// run starts a janitor goroutine that periodically calls cleanup()
// on the target until Stop() is called.
func (j *janitor) run(target cleanupTarget) {
	ticker := time.NewTicker(j.interval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				target.cleanup()
			case <-j.stop:
				return
			}
		}
	}()
}

// stopJanitor stops the janitor goroutine.
func (j *janitor) stopJanitor() {
	select {
	case <-j.stop:
		// already closed, do nothing
	default:
		close(j.stop)
	}
}

// newJanitor creates a janitor with given interval.
func newJanitor(interval time.Duration) *janitor {
	return &janitor{
		interval: interval,
		stop:     make(chan struct{}),
	}
}
