package gbkr

import (
	"errors"
	"net/http"
	"testing"
)

func TestErrorSentinels_Is(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		sentinel error
	}{
		{"ErrBaseURLRequired", ErrBaseURLRequired, ErrBaseURLRequired},
		{"ErrAPIRequest", ErrAPIRequest, ErrAPIRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.sentinel) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.sentinel)
			}
		})
	}
}

func TestAPIError(t *testing.T) {
	err := ErrAPI(404, "404 Not Found", []byte("not found"))

	if !errors.Is(err, ErrAPIRequest) {
		t.Error("APIError should wrap ErrAPIRequest")
	}

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should extract *APIError")
	}
	if apiErr.StatusCode != http.StatusNotFound {
		t.Errorf("StatusCode = %d, want %d", apiErr.StatusCode, http.StatusNotFound)
	}
	if apiErr.Status != "404 Not Found" {
		t.Errorf("Status = %q, want %q", apiErr.Status, "404 Not Found")
	}
	if string(apiErr.Body) != "not found" {
		t.Errorf("Body = %q, want %q", string(apiErr.Body), "not found")
	}
}

func TestErrorStrings(t *testing.T) {
	t.Run("APIError", func(t *testing.T) {
		err := ErrAPI(http.StatusNotFound, "404 Not Found", []byte("oops"))
		msg := err.Error()
		if msg == "" {
			t.Fatal("empty error string")
		}
		// Should contain sentinel text, status, and body.
		for _, want := range []string{"API request failed", "404 Not Found", "oops"} {
			if !contains(msg, want) {
				t.Errorf("error string %q missing %q", msg, want)
			}
		}
	})
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
