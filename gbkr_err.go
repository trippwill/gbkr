package gbkr

import "github.com/trippwill/gbkr/internal/transport"

// Error is a string-based sentinel error type. Use const declarations
// for package-level error values that callers can check with [errors.Is].
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors.
const (
	ErrBaseURLRequired = Error("base URL is required: use WithBaseURL")
	ErrLogoutFailed    = Error("logout was not successful")
)

// TickleError indicates the tickle keepalive was rejected.
type TickleError string

func (e TickleError) Error() string { return "tickle failed: " + string(e) }

// APIError represents a non-2xx response from the IBKR API.
// Re-exported from the internal transport package.
type APIError = transport.APIError

// ErrAPIRequest is a sentinel error for non-2xx API responses.
var ErrAPIRequest error = transport.ErrAPIRequest

// ErrAPI constructs an [APIError] from an HTTP response.
func ErrAPI(statusCode int, status string, body []byte) error {
	return transport.NewAPIError(statusCode, status, body)
}
