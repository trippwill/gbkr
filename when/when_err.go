package when

// Error is the base sentinel type for errors in the when package.
type Error string

func (e Error) Error() string { return string(e) }

const (
	// ErrInvalidDate indicates a date string could not be parsed.
	ErrInvalidDate = Error("invalid date")

	// ErrInvalidDateTime indicates a datetime string could not be parsed.
	ErrInvalidDateTime = Error("invalid datetime")

	// ErrUnsupportedType indicates a SQL Scan received an unsupported Go type.
	ErrUnsupportedType = Error("unsupported type")
)
