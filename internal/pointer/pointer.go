// Package pointer provides utilities for working with pointers - particularly
// handy for libraries that accept everything as pointers.
package pointer

import "fmt"

// Value returns a pointer to the given value.
func Value[V any](v V) *V { return &v }

// Slice takes a slice of values and turns it into a slice of pointers.
func Slice[S []V, V any](s S) []*V {
	slice := make([]*V, len(s))
	for i, v := range s {
		v := v // copy
		slice[i] = &v
	}
	return slice
}

// IfNil returns defaultV if v is nil, otherwise the value of v.
func IfNil[V any](v *V, defaultV V) V {
	if v != nil {
		return *v
	}
	return defaultV
}

type numberType interface {
	~float32 | ~float64 |
		~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

// Float64 returns a pointer to the provided numeric value as a float64.
func Float64[T numberType](v T) *float64 {
	return Value(float64(v))
}

// Stringf is an alias for pointer.Value(fmt.Sprintf)
func Stringf(format string, args ...any) *string {
	return Value(fmt.Sprintf(format, args...))
}
