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
		{"ErrPermissionDenied", ErrPermissionDenied, ErrPermissionDenied},
		{"ErrAPIRequest", ErrAPIRequest, ErrAPIRequest},
		{"ErrUnknownLevel", ErrUnknownLevel, ErrUnknownLevel},
		{"ErrUnknownScope", ErrUnknownScope, ErrUnknownScope},
		{"ErrPermissionsFile", ErrPermissionsFile, ErrPermissionsFile},
		{"ErrPermissionsDecode", ErrPermissionsDecode, ErrPermissionsDecode},
		{"ErrPromptRead", ErrPromptRead, ErrPromptRead},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !errors.Is(tt.err, tt.sentinel) {
				t.Errorf("errors.Is(%v, %v) = false, want true", tt.err, tt.sentinel)
			}
		})
	}
}

func TestPermissionDeniedError(t *testing.T) {
	required := []Permission{{ScopeBrokerage, LevelRead}}
	missing := []Permission{{ScopeBrokerage, LevelRead}}
	err := ErrPermissionsDenied(required, missing)

	if !errors.Is(err, ErrPermissionDenied) {
		t.Error("PermissionDeniedError should wrap ErrPermissionDenied")
	}

	var pde *PermissionDeniedError
	if !errors.As(err, &pde) {
		t.Fatal("errors.As should extract *PermissionDeniedError")
	}
	if len(pde.Required) != 1 || len(pde.Missing) != 1 {
		t.Errorf("got Required=%d, Missing=%d; want 1, 1", len(pde.Required), len(pde.Missing))
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
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

func TestParseError(t *testing.T) {
	err := ErrUnknownLevelValue("bogus")

	if !errors.Is(err, ErrUnknownLevel) {
		t.Error("ParseError should wrap ErrUnknownLevel")
	}

	var pe *ParseError
	if !errors.As(err, &pe) {
		t.Fatal("errors.As should extract *ParseError")
	}
	if pe.Value != "bogus" {
		t.Errorf("Value = %q, want %q", pe.Value, "bogus")
	}
	if pe.Kind != ErrUnknownLevel {
		t.Errorf("Kind = %v, want %v", pe.Kind, ErrUnknownLevel)
	}
}

func TestConfigError(t *testing.T) {
	inner := errors.New("file not found")
	err := ErrPermissionsFileOpen(inner)

	if !errors.Is(err, ErrPermissionsFile) {
		t.Error("ConfigError should wrap ErrPermissionsFile")
	}

	var ce *ConfigError
	if !errors.As(err, &ce) {
		t.Fatal("errors.As should extract *ConfigError")
	}
	if ce.Kind != ErrPermissionsFile {
		t.Errorf("Kind = %v, want %v", ce.Kind, ErrPermissionsFile)
	}
	if !errors.Is(ce.Err, inner) {
		t.Error("inner error not preserved")
	}

	// Multi-unwrap: inner error should be reachable via errors.Is on the original error.
	if !errors.Is(err, inner) {
		t.Error("errors.Is should reach inner error through multi-Unwrap")
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

	t.Run("ParseError", func(t *testing.T) {
		err := ErrUnknownLevelValue("xyz")
		msg := err.Error()
		if !contains(msg, "unknown level") || !contains(msg, "xyz") {
			t.Errorf("ParseError.Error() = %q, want sentinel + value", msg)
		}
	})

	t.Run("ConfigError", func(t *testing.T) {
		err := ErrPermissionsDecoding(errors.New("bad yaml"))
		msg := err.Error()
		if !contains(msg, "decoding permissions") || !contains(msg, "bad yaml") {
			t.Errorf("ConfigError.Error() = %q, want sentinel + inner", msg)
		}
	})

	t.Run("PermissionDeniedError", func(t *testing.T) {
		err := ErrPermissionsDenied(
			[]Permission{{ScopeBrokerage, LevelRead}},
			[]Permission{{ScopeBrokerage, LevelRead}},
		)
		msg := err.Error()
		if !contains(msg, "permission denied") {
			t.Errorf("PermissionDeniedError.Error() = %q, want sentinel text", msg)
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
