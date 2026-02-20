package gbkr

import "fmt"

// Error is a string-based sentinel error type. Use const declarations
// for package-level error values that callers can check with [errors.Is].
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors.
const (
	ErrBaseURLRequired  = Error("base URL is required: use WithBaseURL")
	ErrPermissionDenied = Error("permission denied")
	ErrAPIRequest       = Error("API request failed")
)

// PermissionDeniedError is returned when a capability is requested
// without the required permissions.
type PermissionDeniedError struct {
	Required []Permission
	Missing  []Permission
}

func (e *PermissionDeniedError) Error() string {
	return fmt.Sprintf("%s: missing %v", ErrPermissionDenied, e.Missing)
}

func (e *PermissionDeniedError) Unwrap() error { return ErrPermissionDenied }

// ErrPermissionsDenied constructs a [PermissionDeniedError] listing the
// missing permissions from the required set.
func ErrPermissionsDenied(required, missing []Permission) error {
	return &PermissionDeniedError{Required: required, Missing: missing}
}

// APIError represents a non-2xx response from the IBKR API.
type APIError struct {
	StatusCode int
	Status     string
	Body       []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s: %s", ErrAPIRequest, e.Status, string(e.Body))
}

func (e *APIError) Unwrap() error { return ErrAPIRequest }

// ErrAPI constructs an [APIError] from an HTTP response.
func ErrAPI(statusCode int, status string, body []byte) error {
	return &APIError{StatusCode: statusCode, Status: status, Body: body}
}
