package transport

import (
	"errors"
	"net/http"
	"testing"
)

func TestError_Sentinel(t *testing.T) {
	if ErrAPIRequest.Error() != "API request failed" {
		t.Errorf("ErrAPIRequest = %q", ErrAPIRequest.Error())
	}
}

func TestAPIError_Error(t *testing.T) {
	e := &APIError{StatusCode: 404, Status: "404 Not Found", Body: []byte("oops")}
	got := e.Error()
	want := "API request failed: 404 Not Found: oops"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestAPIError_Unwrap(t *testing.T) {
	e := &APIError{StatusCode: 500}
	if !errors.Is(e, ErrAPIRequest) {
		t.Error("errors.Is should match ErrAPIRequest")
	}
}

func TestNewAPIError(t *testing.T) {
	err := NewAPIError(400, "400 Bad Request", []byte("bad"))
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatal("errors.As should extract *APIError")
	}
	if apiErr.StatusCode != http.StatusBadRequest {
		t.Errorf("StatusCode = %d", apiErr.StatusCode)
	}
}
