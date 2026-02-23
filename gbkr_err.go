package gbkr

import "fmt"

// Error is a string-based sentinel error type. Use const declarations
// for package-level error values that callers can check with [errors.Is].
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors.
const (
	ErrBaseURLRequired = Error("base URL is required: use WithBaseURL")
	ErrAPIRequest      = Error("API request failed")
)

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
