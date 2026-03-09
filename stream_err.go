package gbkr

// Stream sentinel errors.
const (
	ErrStreamNotConnected  = Error("stream is not connected")
	ErrStreamAlreadyClosed = Error("stream is already closed")
)
