package cast

import "unsafe"

// FromStrings casts a slice of strings to a slice of a type whose underlying
// type is string. This is a safe wrapper around the unsafe APIs needed to do
// this in a zero-allocation manner.
//
// NOTE: In the future, Go may make this operation possible
// with simple casts, which would make this function obsolete.
func FromStrings[T ~string](input []string) []T {
	return *(*[]T)(unsafe.Pointer(&input))
}

// ToStrings casts from a slice where the underlying type is a string to a
// slice of strings. This is a safe wrapper around the unsafe APIs needed to do
// this in a zero-allocation manner.
//
// NOTE: In the future, Go may make this operation possible
// with simple casts, which would make this function obsolete.
func ToStrings[T ~string](input []T) []string {
	return *(*[]string)(unsafe.Pointer(&input))
}
