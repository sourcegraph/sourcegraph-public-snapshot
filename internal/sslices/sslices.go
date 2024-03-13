// Package sslices provides generic functions not available in the Go stdlib.
// It's called sslices (Sourcegraph slices) so that it does not conflict with
// the stdlib slices package.
package sslices

// Filter returns a new slice containing only the elements for which the
// provided function returned true.
func Filter[T any](list []T, f func(T) bool) []T {
	filtered := make([]T, 0, len(list))
	for _, v := range list {
		if f(v) {
			filtered = append(filtered, v)
		}
	}

	return filtered
}
