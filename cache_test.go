package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// T034: Test cache cold miss.
func TestCache_ColdMiss(t *testing.T) {
	c := &ttlCache[string]{ttl: time.Minute}
	if got := c.get("k"); got != nil {
		t.Errorf("expected nil on cold cache, got %v", got)
	}
}

// T035: Test cache hit.
func TestCache_Hit(t *testing.T) {
	c := &ttlCache[string]{ttl: time.Minute}
	stored := c.set("k1", "hello")
	got := c.get("k1")
	if got == nil {
		t.Fatal("expected cache hit, got nil")
	}
	if got.Value != "hello" {
		t.Errorf("Value = %q, want %q", got.Value, "hello")
	}
	if got.FetchedAt != stored.FetchedAt {
		t.Errorf("FetchedAt mismatch")
	}
}

// T036: Test cache expiry.
func TestCache_Expiry(t *testing.T) {
	c := &ttlCache[string]{ttl: 50 * time.Millisecond}
	c.set("k", "ephemeral")
	time.Sleep(60 * time.Millisecond)
	if got := c.get("k"); got != nil {
		t.Errorf("expected nil after TTL expiry, got %v", got)
	}
}

// Test cache miss when key changes (different request parameters).
func TestCache_KeyMismatch(t *testing.T) {
	c := &ttlCache[string]{ttl: time.Minute}
	c.set("key-a", "value-a")

	// Same key hits.
	if got := c.get("key-a"); got == nil || got.Value != "value-a" {
		t.Fatalf("expected hit for key-a, got %v", got)
	}

	// Different key misses.
	if got := c.get("key-b"); got != nil {
		t.Errorf("expected miss for key-b, got %v", got)
	}

	// After set with new key, old key misses.
	c.set("key-b", "value-b")
	if got := c.get("key-a"); got != nil {
		t.Errorf("expected miss for key-a after key-b set, got %v", got)
	}
	if got := c.get("key-b"); got == nil || got.Value != "value-b" {
		t.Fatalf("expected hit for key-b, got %v", got)
	}
}
func TestCache_FailClosed(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
				"currency":         "USD",
				"from":             1700000000,
				"to":               1700100000,
				"includesRealTime": true,
				"transactions":     []map[string]any{},
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error")) //nolint:errcheck
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	ar := c.Analysis()

	// Override the cache TTL to something short for the test.
	axr := ar.(*analysisReader)
	axr.txCache.ttl = 50 * time.Millisecond

	// First call — populates cache.
	result, err := ar.Transactions(context.Background(), "U1234567", 265598, 30)
	if err != nil {
		t.Fatalf("first call: %v", err)
	}
	if result.Value.Currency != "USD" {
		t.Errorf("first call Currency = %q", result.Value.Currency)
	}

	// Wait for TTL expiry.
	time.Sleep(60 * time.Millisecond)

	// Second call — server returns 500, should get error.
	_, err = ar.Transactions(context.Background(), "U1234567", 265598, 30)
	if err == nil {
		t.Fatal("expected error on server 500")
	}

	// Third call — cache was invalidated, should try to fetch again (and fail).
	_, err = ar.Transactions(context.Background(), "U1234567", 265598, 30)
	if err == nil {
		t.Fatal("expected error on third call (stale data not returned)")
	}
}

// cacheObserver implements PacingObserver for cache testing.
type cacheObserver struct {
	mu     sync.Mutex
	hits   []string
	misses []string
}

func (o *cacheObserver) OnPacingWait(string, time.Duration) {}
func (o *cacheObserver) OnCacheHit(path string, _ time.Duration) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.hits = append(o.hits, path)
}
func (o *cacheObserver) OnCacheMiss(path string) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.misses = append(o.misses, path)
}

// T038: Test observer on cache.
func TestCache_Observer(t *testing.T) {
	obs := &cacheObserver{}
	c := &ttlCache[string]{
		ttl:      time.Minute,
		observer: obs,
		path:     "/test/cache",
	}

	// Miss on cold cache.
	c.get("k")
	obs.mu.Lock()
	if len(obs.misses) != 1 || obs.misses[0] != "/test/cache" {
		t.Errorf("expected 1 miss for /test/cache, got %v", obs.misses)
	}
	obs.mu.Unlock()

	// Set and hit.
	c.set("k", "value")
	c.get("k")
	obs.mu.Lock()
	if len(obs.hits) != 1 || obs.hits[0] != "/test/cache" {
		t.Errorf("expected 1 hit for /test/cache, got %v", obs.hits)
	}
	obs.mu.Unlock()
}

// T039: Test FetchedAt accuracy.
func TestCache_FetchedAtAccuracy(t *testing.T) {
	c := &ttlCache[int]{ttl: time.Minute}
	before := time.Now()
	stored := c.set("k", 42)
	after := time.Now()

	if stored.FetchedAt.Before(before) || stored.FetchedAt.After(after) {
		t.Errorf("FetchedAt %v not between %v and %v", stored.FetchedAt, before, after)
	}
}
