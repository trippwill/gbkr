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

func TestSession_PermissionDenied(t *testing.T) {
	c, err := NewClient(
		WithBaseURL("http://localhost"),
		WithPermissions(PermissionSet{}),
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Session(c)
	if !errors.Is(err, ErrPermissionDenied) {
		t.Errorf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestSession_InitBrokerageSession(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}
	sc, err := Session(c)
	if err != nil {
		t.Fatal(err)
	}

	compete := true
	result, err := sc.InitBrokerageSession(context.Background(), &models.SSOInitRequest{Compete: &compete})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Authenticated {
		t.Error("expected Authenticated=true")
	}
	if result.ServerName != "srv1" {
		t.Errorf("ServerName = %q, want %q", result.ServerName, "srv1")
	}
}

func TestSession_SessionStatus(t *testing.T) {
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

	c, err := NewClient(WithBaseURL(srv.URL), WithPermissions(AllPermissions()))
	if err != nil {
		t.Fatal(err)
	}
	sc, err := Session(c)
	if err != nil {
		t.Fatal(err)
	}

	result, err := sc.SessionStatus(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Connected {
		t.Error("expected Connected=true")
	}
}
