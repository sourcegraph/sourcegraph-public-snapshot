package pointers

// Ptr returns a pointer to any value.
// Useful in tests or when pointer without a variable is needed.
func Ptr[T any](val T) *T {
	return &val
}

// NonZeroPtr returns nil for zero value, otherwise pointer to value
func NonZeroPtr[T comparable](val T) *T {
	var zero T
	if val == zero {
		return nil
	}
	return Ptr(val)
}

// Deref safely dereferences a pointer. If pointer is nil, returns defaultValue,
// otherwise returns dereferenced value.
func Deref[T any](v *T, defaultValue T) T {
	if v != nil {
		return *v
	}

	return defaultValue
}
