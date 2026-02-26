package transport

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

func newTestTransport(t *testing.T, handler http.HandlerFunc) (*Transport, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	tr := &Transport{
		BaseURL:    srv.URL,
		HTTPClient: http.DefaultClient,
		Logger:     slog.Default(),
	}
	return tr, srv
}

func TestGet_Success(t *testing.T) {
	type payload struct {
		Name string `json:"name"`
		N    int    `json:"n"`
	}

	tr, srv := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(payload{Name: "test", N: 42}) //nolint:errcheck
	})
	defer srv.Close()

	var result payload
	err := tr.Get(context.Background(), "/test", nil, &result)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if result.Name != "test" || result.N != 42 {
		t.Errorf("got %+v, want {Name:test N:42}", result)
	}
}

func TestGet_QueryParams(t *testing.T) {
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		if got := r.URL.Query().Get("foo"); got != "bar" {
			t.Errorf("query foo = %q, want %q", got, "bar")
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	q := url.Values{"foo": {"bar"}}
	err := tr.Get(context.Background(), "/test", q, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
}

func TestGet_NonOK(t *testing.T) {
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found")) //nolint:errcheck
	})
	defer srv.Close()

	err := tr.Get(context.Background(), "/missing", nil, nil)
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

func TestPost_Success(t *testing.T) {
	type req struct {
		Value string `json:"value"`
	}
	type resp struct {
		OK bool `json:"ok"`
	}

	tr, srv := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
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
	err := tr.Post(context.Background(), "/action", req{Value: "hello"}, &result)
	if err != nil {
		t.Fatalf("Post: %v", err)
	}
	if !result.OK {
		t.Error("expected OK=true")
	}
}

func TestPost_NilBody(t *testing.T) {
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer srv.Close()

	err := tr.Post(context.Background(), "/action", nil, nil)
	if err != nil {
		t.Fatalf("Post nil body: %v", err)
	}
}

func TestPost_NonOK(t *testing.T) {
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("server error")) //nolint:errcheck
	})
	defer srv.Close()

	err := tr.Post(context.Background(), "/fail", nil, nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}
	if !errors.Is(err, ErrAPIRequest) {
		t.Errorf("expected ErrAPIRequest, got %v", err)
	}
}

type mockHook struct {
	mu       sync.Mutex
	before   []string
	after    []string
	beforeFn func(ctx context.Context, method, path string) error
}

func (h *mockHook) BeforeRequest(ctx context.Context, method, path string) error {
	h.mu.Lock()
	h.before = append(h.before, method+" "+path)
	h.mu.Unlock()
	if h.beforeFn != nil {
		return h.beforeFn(ctx, method, path)
	}
	return nil
}

func (h *mockHook) AfterRequest(method, path string) {
	h.mu.Lock()
	h.after = append(h.after, method+" "+path)
	h.mu.Unlock()
}

func TestGet_WithHook(t *testing.T) {
	hook := &mockHook{}
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	tr.Hook = hook
	defer srv.Close()

	err := tr.Get(context.Background(), "/hooked", nil, nil)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if len(hook.before) != 1 || hook.before[0] != "GET /hooked" {
		t.Errorf("before = %v", hook.before)
	}
	if len(hook.after) != 1 || hook.after[0] != "GET /hooked" {
		t.Errorf("after = %v", hook.after)
	}
}

func TestGet_HookError(t *testing.T) {
	errPacing := errors.New("rate limited")
	hook := &mockHook{
		beforeFn: func(_ context.Context, _, _ string) error {
			return errPacing
		},
	}
	tr, srv := newTestTransport(t, func(w http.ResponseWriter, _ *http.Request) {
		t.Error("handler should not be called when hook returns error")
		w.WriteHeader(http.StatusOK)
	})
	tr.Hook = hook
	defer srv.Close()

	err := tr.Get(context.Background(), "/blocked", nil, nil)
	if !errors.Is(err, errPacing) {
		t.Errorf("expected errPacing, got %v", err)
	}
}

// captureHandler records slog records for assertion.
type captureHandler struct {
	mu      sync.Mutex
	records []slog.Record
}

func (h *captureHandler) Enabled(_ context.Context, _ slog.Level) bool { return true }
func (h *captureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	h.records = append(h.records, r)
	h.mu.Unlock()
	return nil
}
func (h *captureHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *captureHandler) WithGroup(_ string) slog.Handler      { return h }

func TestEmitOp_Success(t *testing.T) {
	ch := &captureHandler{}
	tr := &Transport{Logger: slog.New(ch)}

	tr.EmitOp(context.Background(), "test-op", nil, 100*time.Millisecond)

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	r := ch.records[0]
	if r.Level != slog.LevelInfo {
		t.Errorf("level = %v, want Info", r.Level)
	}
	var gotOp string
	r.Attrs(func(a slog.Attr) bool {
		if a.Key == "op" {
			gotOp = a.Value.String()
		}
		return true
	})
	if gotOp != "test-op" {
		t.Errorf("op = %q, want %q", gotOp, "test-op")
	}
}

func TestEmitOp_Error(t *testing.T) {
	ch := &captureHandler{}
	tr := &Transport{Logger: slog.New(ch)}

	tr.EmitOp(context.Background(), "fail-op", errors.New("boom"), 50*time.Millisecond)

	if len(ch.records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(ch.records))
	}
	r := ch.records[0]
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
	if gotErr != "boom" {
		t.Errorf("error attr = %q, want %q", gotErr, "boom")
	}
}
