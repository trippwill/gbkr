package brokerage

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippwill/gbkr"
)

func newTestBrokerageClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := gbkr.NewClient(gbkr.WithBaseURL(srv.URL), gbkr.WithRateLimit(nil))
	if err != nil {
		srv.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return &Client{client: c}, srv
}

func TestNewSession(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/auth/ssodh/init" {
			t.Errorf("path = %q, want /iserver/auth/ssodh/init", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"authenticated": true,
			"connected":     true,
		})
	}))
	defer srv.Close()

	c, err := gbkr.NewClient(gbkr.WithBaseURL(srv.URL), gbkr.WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	bc, err := NewSession(context.Background(), c, &SSOInitRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if bc == nil {
		t.Fatal("expected non-nil Client")
	}
}

func TestNewSession_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("unauthorized")) //nolint:errcheck
	}))
	defer srv.Close()

	c, err := gbkr.NewClient(gbkr.WithBaseURL(srv.URL), gbkr.WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = NewSession(context.Background(), c, &SSOInitRequest{})
	if err == nil {
		t.Fatal("expected error for 401")
	}
}
