package brokerage

import "fmt"

// Error is a string-based sentinel error type for the brokerage package.
type Error string

func (e Error) Error() string { return string(e) }

// Sentinel errors.
const (
	ErrInvalidCount = Error("count must be positive")
	ErrInvalidUnit  = Error("invalid unit")
)

// ValidationError is returned when a BarSize or TimePeriod parameter is invalid.
type ValidationError struct {
	Kind  Error
	Value int
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %d", e.Kind, e.Value)
}

func (e *ValidationError) Unwrap() error { return e.Kind }

// ErrInvalidCountValue constructs a [ValidationError] for a non-positive count.
func ErrInvalidCountValue(count int) error {
	return &ValidationError{Kind: ErrInvalidCount, Value: count}
}

// ErrInvalidUnitValue constructs a [ValidationError] for an unrecognized unit.
func ErrInvalidUnitValue(unit int) error {
	return &ValidationError{Kind: ErrInvalidUnit, Value: unit}
}
