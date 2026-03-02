package gbkr

import (
	"context"
	"encoding/json"
	"errors"
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

func TestLogout(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/logout" {
				t.Errorf("path = %q, want /logout", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": true}) //nolint:errcheck
		}))
		defer srv.Close()

		c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
		if err != nil {
			t.Fatal(err)
		}

		if err := c.Logout(context.Background()); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("status_false", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": false}) //nolint:errcheck
		}))
		defer srv.Close()

		c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
		if err != nil {
			t.Fatal(err)
		}

		err = c.Logout(context.Background())
		if !errors.Is(err, ErrLogoutFailed) {
			t.Errorf("got %v, want ErrLogoutFailed", err)
		}
	})
}

func TestReauthenticate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/iserver/reauthenticate" {
			t.Errorf("path = %q, want /iserver/reauthenticate", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %q, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{ //nolint:errcheck
			"authenticated": true,
			"connected":     true,
			"competing":     false,
		})
	}))
	defer srv.Close()

	c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
	if err != nil {
		t.Fatal(err)
	}

	result, err := c.Reauthenticate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !result.Authenticated {
		t.Error("expected Authenticated=true")
	}
	if !result.Connected {
		t.Error("expected Connected=true")
	}
}

func TestTickle(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/tickle" {
				t.Errorf("path = %q, want /tickle", r.URL.Path)
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"ssoExpires": 300000}) //nolint:errcheck
		}))
		defer srv.Close()

		c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
		if err != nil {
			t.Fatal(err)
		}

		expires, err := c.Tickle(context.Background())
		if err != nil {
			t.Fatal(err)
		}
		if expires != 300000 {
			t.Errorf("ssoExpires = %d, want 300000", expires)
		}
	})

	t.Run("error_response", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"error": "session expired"}) //nolint:errcheck
		}))
		defer srv.Close()

		c, err := NewClient(WithBaseURL(srv.URL), WithRateLimit(nil))
		if err != nil {
			t.Fatal(err)
		}

		_, err = c.Tickle(context.Background())
		if err == nil {
			t.Fatal("expected error")
		}
		var tickleErr TickleError
		if !errors.As(err, &tickleErr) {
			t.Errorf("got %T, want TickleError", err)
		}
	})
}
