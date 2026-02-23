package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/trippwill/gbkr/models"
)

func TestSessionStatus_NoPermissionRequired(t *testing.T) {
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

	// No permissions — SessionStatus is ungated.
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

func TestBrokerageSession_PermissionDenied(t *testing.T) {
	c, err := NewClient(WithBaseURL("http://localhost"), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	_, err = c.BrokerageSession(context.Background(), &models.SSOInitRequest{})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestBrokerageSession_InitBrokerageSession(t *testing.T) {
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
			"serverName":    "srv1",
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(ReadOnly()), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	compete := true
	bc, err := c.BrokerageSession(context.Background(), &models.SSOInitRequest{Compete: &compete})
	if err != nil {
		t.Fatal(err)
	}
	if bc == nil {
		t.Fatal("expected non-nil BrokerageClient")
	}
	if bc.Client != c {
		t.Error("expected BrokerageClient to embed the original Client")
	}
}
