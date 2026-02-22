package gbkr

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func testPacingPolicy(path, method string, interval time.Duration, burst int) *PacingPolicy {
	return &PacingPolicy{
		global: rate.NewLimiter(rate.Limit(1000), 100), // effectively unlimited
		rules: []pathLimiter{{
			prefix:      path,
			method:      method,
			rateLimiter: rate.NewLimiter(rate.Every(interval), burst),
		}},
	}
}

func testSemaphorePolicy(path, method string, maxConcurrent int) *PacingPolicy {
	return &PacingPolicy{
		global: rate.NewLimiter(rate.Limit(1000), 100),
		rules: []pathLimiter{{
			prefix:    path,
			method:    method,
			semaphore: make(chan struct{}, maxConcurrent),
		}},
	}
}

// T015: Test per-path rate enforcement.
func TestPacing_PerPathRate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	policy := testPacingPolicy("/test/path", "GET", 100*time.Millisecond, 1)
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(policy),
	)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for i := range 5 {
		if err := c.doGet(context.Background(), "/test/path", nil, nil); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	if elapsed < 400*time.Millisecond {
		t.Errorf("elapsed %v, want >= 400ms", elapsed)
	}
}

// T016: Test global ceiling.
func TestPacing_GlobalCeiling(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	policy := &PacingPolicy{
		global: rate.NewLimiter(rate.Limit(10), 1), // 10/s = 100ms between, burst 1
		rules:  nil,
	}
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(policy),
	)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for i := range 5 {
		if err := c.doGet(context.Background(), "/random/path", nil, nil); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	if elapsed < 350*time.Millisecond {
		t.Errorf("elapsed %v, want >= ~400ms", elapsed)
	}
}

// T017: Test concurrency semaphore.
func TestPacing_ConcurrencySemaphore(t *testing.T) {
	var inFlight atomic.Int32
	var maxInFlight atomic.Int32

	unblock := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		cur := inFlight.Add(1)
		for {
			old := maxInFlight.Load()
			if cur <= old {
				break
			}
			if maxInFlight.CompareAndSwap(old, cur) {
				break
			}
		}
		<-unblock
		inFlight.Add(-1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	policy := testSemaphorePolicy("/iserver/marketdata/history", "GET", 2)
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(policy),
	)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	for range 4 {
		wg.Go(func() {
			c.doGet(context.Background(), "/iserver/marketdata/history", nil, nil) //nolint:errcheck
		})
	}

	// Give goroutines time to start and hit the semaphore.
	time.Sleep(100 * time.Millisecond)

	if got := maxInFlight.Load(); got > 2 {
		t.Errorf("max in-flight = %d, want <= 2", got)
	}

	close(unblock)
	wg.Wait()
}

// T018: Test context cancellation.
func TestPacing_ContextCancellation(t *testing.T) {
	var requestCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	// Slow policy: 10s interval, burst 1
	policy := testPacingPolicy("/test", "GET", 10*time.Second, 1)
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(policy),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Consume the burst.
	if err := c.doGet(context.Background(), "/test", nil, nil); err != nil {
		t.Fatalf("first request: %v", err)
	}
	countAfterFirst := requestCount.Load()

	// Fire with already-cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = c.doGet(ctx, "/test", nil, nil)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}

	// No additional HTTP request should have been made.
	if got := requestCount.Load(); got != countAfterFirst {
		t.Errorf("request count = %d, want %d (no extra request)", got, countAfterFirst)
	}
}

// mockObserver implements PacingObserver for testing.
type mockObserver struct {
	mu    sync.Mutex
	waits []struct {
		path string
		dur  time.Duration
	}
}

func (m *mockObserver) OnPacingWait(path string, d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.waits = append(m.waits, struct {
		path string
		dur  time.Duration
	}{path, d})
}

func (m *mockObserver) OnCacheHit(string, time.Duration) {}
func (m *mockObserver) OnCacheMiss(string)               {}

// T019: Test PacingObserver.
func TestPacing_Observer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	obs := &mockObserver{}
	policy := testPacingPolicy("/test/obs", "GET", 100*time.Millisecond, 1)
	policy.observer = obs

	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(policy),
	)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.doGet(context.Background(), "/test/obs", nil, nil); err != nil {
		t.Fatal(err)
	}

	obs.mu.Lock()
	defer obs.mu.Unlock()
	if len(obs.waits) != 1 {
		t.Fatalf("observer waits = %d, want 1", len(obs.waits))
	}
	if obs.waits[0].path != "/test/obs" {
		t.Errorf("observer path = %q, want /test/obs", obs.waits[0].path)
	}
}

// T023: Test WithRateLimit(nil) disables pacing.
func TestPacing_WithRateLimitNil(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}

	start := time.Now()
	for i := range 20 {
		if err := c.doGet(context.Background(), "/pa/transactions", nil, nil); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)

	if elapsed >= time.Second {
		t.Errorf("elapsed %v, want < 1s (pacing disabled)", elapsed)
	}
}

// T024: Test custom policy overrides defaults.
func TestPacing_CustomPolicyOverridesDefaults(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	customPolicy := &PacingPolicy{
		global: rate.NewLimiter(rate.Limit(1000), 100), // effectively unlimited
		rules: []pathLimiter{{
			prefix:      "/test",
			method:      "GET",
			rateLimiter: rate.NewLimiter(rate.Every(500*time.Millisecond), 1),
		}},
	}
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(customPolicy),
	)
	if err != nil {
		t.Fatal(err)
	}

	// Fire 3 requests to /test — should take >= 1000ms (2 waits of 500ms).
	start := time.Now()
	for i := range 3 {
		if err := c.doGet(context.Background(), "/test", nil, nil); err != nil {
			t.Fatalf("request %d: %v", i, err)
		}
	}
	elapsed := time.Since(start)
	if elapsed < time.Second {
		t.Errorf("strict path elapsed %v, want >= 1s", elapsed)
	}

	// Requests to an unlisted path should use the unlimited global.
	start = time.Now()
	for i := range 5 {
		if err := c.doGet(context.Background(), "/other/path", nil, nil); err != nil {
			t.Fatalf("other request %d: %v", i, err)
		}
	}
	elapsed = time.Since(start)
	if elapsed >= 500*time.Millisecond {
		t.Errorf("unlisted path elapsed %v, want < 500ms", elapsed)
	}
}
