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

func TestWithBaseURL_Validation(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"http", "http://localhost:5000/v1/api", false},
		{"https", "https://gateway.example.com/v1/api", false},
		{"no scheme", "localhost:5000", true},
		{"ftp scheme", "ftp://localhost", true},
		{"empty", "", true},
		{"no host", "http://", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewClient(WithBaseURL(tt.url), WithRateLimit(nil))
			if (err != nil) != tt.wantErr {
				t.Errorf("WithBaseURL(%q): err = %v, wantErr = %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
