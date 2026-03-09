package transport

// WebSocket sentinel errors.
const (
	ErrWSClosed = Error("WebSocket connection closed")
	ErrWSSend   = Error("WebSocket send failed")
)
