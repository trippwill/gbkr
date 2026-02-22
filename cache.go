package gbkr

import (
	"sync"
	"time"
)

// Cached wraps a value with its fetch timestamp.
type Cached[T any] struct {
	Value     T
	FetchedAt time.Time
}

type ttlCache[T any] struct {
	mu       sync.Mutex
	entry    *Cached[T]
	ttl      time.Duration
	observer PacingObserver
	path     string // for observer notifications
}

func (c *ttlCache[T]) get() *Cached[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.entry == nil {
		if c.observer != nil {
			c.observer.OnCacheMiss(c.path)
		}
		return nil
	}
	age := time.Since(c.entry.FetchedAt)
	if age >= c.ttl {
		c.entry = nil // expired
		if c.observer != nil {
			c.observer.OnCacheMiss(c.path)
		}
		return nil
	}
	if c.observer != nil {
		c.observer.OnCacheHit(c.path, age)
	}
	return c.entry
}

func (c *ttlCache[T]) set(value T) *Cached[T] {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entry = &Cached[T]{Value: value, FetchedAt: time.Now()}
	return c.entry
}

func (c *ttlCache[T]) invalidate() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entry = nil
}
