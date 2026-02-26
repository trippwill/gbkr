package transport

import "fmt"

// Error is a string-based sentinel error type for transport errors.
type Error string

func (e Error) Error() string { return string(e) }

// ErrAPIRequest is a sentinel for non-2xx API responses.
const ErrAPIRequest = Error("API request failed")

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

// NewAPIError constructs an [APIError] from HTTP response data.
func NewAPIError(statusCode int, status string, body []byte) error {
	return &APIError{StatusCode: statusCode, Status: status, Body: body}
}
