package gbkr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/auth/status" {
			t.Errorf("path = %q, want /iserver/auth/status", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"authenticated": true,
			"connected":     true,
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.SessionStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Connected {
		t.Error("expected Connected=true")
	}
}
