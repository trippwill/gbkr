package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RequestHook intercepts HTTP requests for pacing or other cross-cutting concerns.
type RequestHook interface {
	BeforeRequest(ctx context.Context, method, path string) error
	AfterRequest(method, path string)
}

// Transport provides shared HTTP plumbing for IBKR API requests.
type Transport struct {
	BaseURL    string
	HTTPClient *http.Client
	Logger     *slog.Logger
	Hook       RequestHook
}

// Get performs an HTTP GET request.
func (t *Transport) Get(ctx context.Context, path string, query url.Values, result any) error {
	u, err := url.JoinPath(t.BaseURL, path)
	if err != nil {
		return fmt.Errorf("building request URL: %w", err)
	}
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	return t.doRequest(req, result)
}

// Post performs an HTTP POST request.
func (t *Transport) Post(ctx context.Context, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshaling request body: %w", err)
		}
		bodyReader = strings.NewReader(string(data))
	}

	u, err := url.JoinPath(t.BaseURL, path)
	if err != nil {
		return fmt.Errorf("building request URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bodyReader)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return t.doRequest(req, result)
}

func (t *Transport) doRequest(req *http.Request, result any) error {
	if t.Hook != nil {
		path := req.URL.Path
		if err := t.Hook.BeforeRequest(req.Context(), req.Method, path); err != nil {
			return err
		}
		defer t.Hook.AfterRequest(req.Method, path)
	}

	//nolint:gosec // G704: BaseURL is validated at client construction; user-configured gateway is by design.
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return NewAPIError(resp.StatusCode, resp.Status, data)
	}

	if result != nil && len(data) > 0 {
		if err := json.Unmarshal(data, result); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// EmitOp logs a structured operation event.
// Successful operations emit at LevelInfo; failures emit at LevelWarn.
func (t *Transport) EmitOp(ctx context.Context, op string, err error, dur time.Duration, attrs ...slog.Attr) {
	level := slog.LevelInfo
	if err != nil {
		level = slog.LevelWarn
	}
	allAttrs := make([]slog.Attr, 0, 3+len(attrs))
	allAttrs = append(allAttrs, slog.String("op", op))
	allAttrs = append(allAttrs, slog.Duration("duration", dur))
	if err != nil {
		allAttrs = append(allAttrs, slog.String("error", err.Error()))
	}
	allAttrs = append(allAttrs, attrs...)
	t.Logger.LogAttrs(ctx, level, "operation", allAttrs...)
}
