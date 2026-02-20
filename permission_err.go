package gbkr

import "fmt"

// Sentinel errors for permission parsing.
const (
	ErrUnknownArea     = Error("unknown area")
	ErrUnknownResource = Error("unknown resource")
	ErrUnknownAction   = Error("unknown action")
)

// ParseError is returned when a permission field string cannot be mapped
// to a known enum value.
type ParseError struct {
	Kind  Error
	Value string
}

func (e *ParseError) Error() string {
	return fmt.Sprintf("%s: %q", e.Kind, e.Value)
}

func (e *ParseError) Unwrap() error { return e.Kind }

// ErrUnknownAreaValue constructs a [ParseError] for an unrecognized area string.
func ErrUnknownAreaValue(value string) error {
	return &ParseError{Kind: ErrUnknownArea, Value: value}
}

// ErrUnknownResourceValue constructs a [ParseError] for an unrecognized resource string.
func ErrUnknownResourceValue(value string) error {
	return &ParseError{Kind: ErrUnknownResource, Value: value}
}

// ErrUnknownActionValue constructs a [ParseError] for an unrecognized action string.
func ErrUnknownActionValue(value string) error {
	return &ParseError{Kind: ErrUnknownAction, Value: value}
}
