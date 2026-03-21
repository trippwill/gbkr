package flex

import (
	"errors"
	"fmt"
)

// Error is a string-based sentinel error type. Use const declarations
// for package-level error values that callers can check with [errors.Is].
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors.
const (
	// ErrTokenExpired indicates the Flex API token has expired (IBKR error code 1012).
	ErrTokenExpired = Error("flex token expired")

	// ErrRateLimited indicates the request was rate-limited (IBKR error code 1018).
	ErrRateLimited = Error("flex rate limit exceeded")

	// ErrQueryNotFound indicates the query ID was not found (IBKR error code 1013).
	ErrQueryNotFound = Error("flex query not found")

	// ErrReportNotReady indicates the report is still being generated (IBKR error code 1019).
	ErrReportNotReady = Error("flex report not ready")

	// ErrInvalidToken indicates the token is malformed or rejected.
	ErrInvalidToken = Error("flex token invalid")

	// ErrServiceUnavailable indicates the Flex Web Service is temporarily unavailable.
	ErrServiceUnavailable = Error("flex service unavailable")

	// ErrWrongFormat indicates the Flex query is not configured for XML output.
	// To fix: edit the query in IBKR Client Portal → Flex Queries and set Format to XML.
	ErrWrongFormat = Error("flex query is returning CSV; change Format to XML in the IBKR Flex query template")
)

// ResponseError wraps an error code and message from a FlexStatementResponse.
type ResponseError struct {
	Code    int
	Message string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("flex response error %d: %s", e.Code, e.Message)
}

// Unwrap returns the appropriate sentinel error for known IBKR error codes,
// or nil for unrecognized codes.
func (e *ResponseError) Unwrap() error {
	switch e.Code {
	case 1012:
		return ErrTokenExpired
	case 1013:
		return ErrQueryNotFound
	case 1018:
		return ErrRateLimited
	case 1019:
		return ErrReportNotReady
	default:
		return nil
	}
}

// ErrResponse constructs a [ResponseError] from an IBKR error code and message.
func ErrResponse(code int, message string) error {
	return &ResponseError{Code: code, Message: message}
}

// IsRetryable reports whether the error indicates the request can be retried.
func IsRetryable(err error) bool {
	return errors.Is(err, ErrReportNotReady) || errors.Is(err, ErrRateLimited)
}
