package gbkr

import "fmt"

// Sentinel errors for permission parsing.
const (
	ErrUnknownLevel = Error("unknown level")
	ErrUnknownScope = Error("unknown scope")
)

// ParseError is returned when a permission field string cannot be mapped
// to a known value.
type ParseError struct {
	Kind  Error
	Value string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %q", e.Kind, e.Value)
}

func (e *ParseError) Unwrap() error { return e.Kind }

// ErrUnknownLevelValue constructs a [ParseError] for an unrecognized level string.
func ErrUnknownLevelValue(value string) error {
	return &ParseError{Kind: ErrUnknownLevel, Value: value}
}

// ErrUnknownScopeValue constructs a [ParseError] for an unrecognized scope string.
func ErrUnknownScopeValue(value string) error {
	return &ParseError{Kind: ErrUnknownScope, Value: value}
}
