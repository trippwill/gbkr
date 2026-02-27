package gbkr

import (
	"log/slog"
	"net/http"
	"testing"
)

func TestWithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithHTTPClient(custom),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.t.HTTPClient != custom {
		t.Error("WithHTTPClient did not set the client")
	}
}

func TestWithLogger(t *testing.T) {
	h := newCaptureHandler()
	custom := slog.New(h)
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithLogger(custom),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.t.Logger == nil {
		t.Fatal("logger should not be nil")
	}
	// Logger should have gbkr group applied (set in NewClient).
	// Verify by emitting and checking the handler received it.
	c.t.Logger.Info("test")
	if h.count() != 1 {
		t.Errorf("expected 1 record, got %d", h.count())
	}
}

func TestWithLogger_DefaultWhenNil(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithRateLimit(nil),
	)
	if err != nil {
		t.Fatal(err)
	}
	if c.t.Logger == nil {
		t.Error("default logger should not be nil")
	}
}
