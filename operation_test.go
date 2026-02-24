package gbkr

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"testing"
	"time"
)

// captureHandler is a slog.Handler that records all log records for test assertions.
type captureHandler struct {
	mu      *sync.Mutex
	records *[]slog.Record
	groups  []string
	attrs   []slog.Attr
}

func newCaptureHandler() *captureHandler {
	return &captureHandler{
		mu:      &sync.Mutex{},
		records: &[]slog.Record{},
	}
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }

func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	*h.records = append(*h.records, r)
	return nil
}

func (h *captureHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &captureHandler{mu: h.mu, records: h.records, groups: h.groups, attrs: append(h.attrs, attrs...)}
}

func (h *captureHandler) WithGroup(name string) slog.Handler {
	return &captureHandler{mu: h.mu, records: h.records, groups: append(h.groups, name), attrs: h.attrs}
}

func (h *captureHandler) last() slog.Record {
	h.mu.Lock()
	defer h.mu.Unlock()
	return (*h.records)[len(*h.records)-1]
}

func (h *captureHandler) count() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(*h.records)
}

func TestEmitOp_Success(t *testing.T) {
	h := newCaptureHandler()
	logger := slog.New(h)

	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithRateLimit(nil),
		WithLogger(logger),
	)
	if err != nil {
		t.Fatal(err)
	}

	c.emitOp(context.Background(), OpListAccounts, nil, 42*time.Millisecond)

	if h.count() != 1 {
		t.Fatalf("got %d records, want 1", h.count())
	}

	r := h.last()
	if r.Level != slog.LevelInfo {
		t.Errorf("level = %v, want Info", r.Level)
	}
	if r.Message != "operation" {
		t.Errorf("message = %q, want %q", r.Message, "operation")
	}

	var gotOp, gotErr string
	var gotDur time.Duration
	r.Attrs(func(a slog.Attr) bool {
		switch a.Key {
		case "op":
			gotOp = a.Value.String()
		case "duration":
			gotDur = a.Value.Duration()
		case "error":
			gotErr = a.Value.String()
		}
		return true
	})

	if gotOp != string(OpListAccounts) {
		t.Errorf("op = %q, want %q", gotOp, OpListAccounts)
	}
	if gotDur != 42*time.Millisecond {
		t.Errorf("duration = %v, want 42ms", gotDur)
	}
	if gotErr != "" {
		t.Errorf("error attr should be absent, got %q", gotErr)
	}
}

func TestEmitOp_Error(t *testing.T) {
	h := newCaptureHandler()
	logger := slog.New(h)

	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithRateLimit(nil),
		WithLogger(logger),
	)
	if err != nil {
		t.Fatal(err)
	}

	testErr := errors.New("connection refused")
	c.emitOp(context.Background(), OpSessionStatus, testErr, 5*time.Millisecond)

	r := h.last()
	if r.Level != slog.LevelWarn {
		t.Errorf("level = %v, want Warn", r.Level)
	}

	var gotErr string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "error" {
			gotErr = a.Value.String()
		}
		return true
	})
	if gotErr != "connection refused" {
		t.Errorf("error = %q, want %q", gotErr, "connection refused")
	}
}

func TestEmitOp_ContextAttrs(t *testing.T) {
	h := newCaptureHandler()
	logger := slog.New(h)

	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithRateLimit(nil),
		WithLogger(logger),
	)
	if err != nil {
		t.Fatal(err)
	}

	c.emitOp(context.Background(), OpPortfolioPosition, nil, time.Millisecond,
		slog.String("account_id", "U1234567"),
		slog.Int64("conid", 265598),
	)

	r := h.last()
	attrs := map[string]slog.Value{}
	r.Attrs(func(a slog.Attr) bool {
		attrs[a.Key] = a.Value
		return true
	})

	if v, ok := attrs["account_id"]; !ok || v.String() != "U1234567" {
		t.Errorf("account_id = %v, want U1234567", attrs["account_id"])
	}
	if v, ok := attrs["conid"]; !ok || v.Int64() != 265598 {
		t.Errorf("conid = %v, want 265598", attrs["conid"])
	}
}
