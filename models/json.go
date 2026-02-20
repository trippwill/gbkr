package models

// deref safely dereferences a pointer, returning the zero value if nil.
func deref[T any](p *T) T {
	if p == nil {
		var zero T
		return zero
	}
	return *p
}
