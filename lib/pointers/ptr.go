package pointers

import "fmt"

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

// Deref safely dereferences a pointer. If pointer is nil, returns default value,
// otherwise returns dereferenced value.
func Deref[T any](v *T, defaultValue T) T {
	if v != nil {
		return *v
	}

	return defaultValue
}

// Deref safely dereferences a pointer. If pointer is nil, it returns a zero value,
// otherwise returns dereferenced value.
func DerefZero[T any](v *T) T {
	if v != nil {
		return *v
	}

	var defaultValue T
	return defaultValue
}

type numberType interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Float64 returns a pointer to the provided numeric value as a float64.
func Float64[T numberType](v T) *float64 {
	return Ptr(float64(v))
}

// Stringf is an alias for Ptr(fmt.Sprintf(format, a...))
func Stringf(format string, a ...any) *string {
	return Ptr(fmt.Sprintf(format, a...))
}

// Slice takes a slice of values and turns it into a slice of pointers.
func Slice[S []V, V any](s S) []*V {
	slice := make([]*V, len(s))
	for i, v := range s {
		v := v // copy
		slice[i] = &v
	}
	return slice
}
