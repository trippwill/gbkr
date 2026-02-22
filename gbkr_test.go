package gbkr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	c, err := NewClient(
		WithBaseURL(srv.URL),
		WithPermissions(AllPermissions()),
		WithRateLimit(nil),
	)
	if err != nil {
		srv.Close()
		t.Fatalf("NewClient: %v", err)
	}
	return c, srv
}

func TestDoGet_Success(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload{Name: "test", N: 42}) //nolint:errcheck
	})
	defer srv.Close()

	var result payload
	err := c.doGet(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
	if result.Name != "test" || result.N != 42 {
		t.Errorf("got %+v, want {Name:test N:42}", result)
	}
}

func TestDoGet_QueryParams(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("foo"); got != "bar" {
			t.Errorf("query foo = %q, want %q", got, "bar")
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	q := url.Values{"foo": {"bar"}}
	err := c.doGet(context.Background(), "/test", q, nil)
	if err != nil {
		t.Fatalf("doGet: %v", err)
	}
}

func TestDoGet_NonOK(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found")) //nolint:errcheck
	})
	defer srv.Close()

	err := c.doGet(context.Background(), "/missing", nil, nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !errors.Is(err, ErrAPIRequest) {
		t.Errorf("expected ErrAPIRequest, got %v", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should extract *APIError")
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
}

func TestDoPost_Success(t *testing.T) {
	type req struct {
		Value string `json:"value"`
	}
	type resp struct {
		OK bool `json:"ok"`
	}

	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("Content-Type = %q, want application/json", ct)
		}

		var body req
		json.NewDecoder(r.Body).Decode(&body) //nolint:errcheck
		if body.Value != "hello" {
			t.Errorf("body.Value = %q, want %q", body.Value, "hello")
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp{OK: true}) //nolint:errcheck
	})
	defer srv.Close()

	var result resp
	err := c.doPost(context.Background(), "/action", req{Value: "hello"}, &result)
	if err != nil {
		t.Fatalf("doPost: %v", err)
	}
	if !result.OK {
		t.Error("expected OK=true")
	}
}

func TestDoPost_NilBody(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	err := c.doPost(context.Background(), "/action", nil, nil)
	if err != nil {
		t.Fatalf("doPost nil body: %v", err)
	}
}

func TestDoPost_NonOK(t *testing.T) {
	c, srv := newTestClient(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error")) //nolint:errcheck
	})
	defer srv.Close()

	err := c.doPost(context.Background(), "/fail", nil, nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !errors.Is(err, ErrAPIRequest) {
		t.Errorf("expected ErrAPIRequest, got %v", err)
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should extract *APIError")
	}
	if apiErr.StatusCode != http.StatusInternalServerError {
		t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
	}
}

func TestNewClient_NoBaseURL(t *testing.T) {
	_, err := NewClient()
	if err == nil {
		t.Fatal("expected error without base URL")
	}
	if !errors.Is(err, ErrBaseURLRequired) {
		t.Errorf("expected ErrBaseURLRequired, got %v", err)
	}
}
