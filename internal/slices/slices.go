package slices

import (
	"golang.org/x/exp/constraints"
)

// Min returns the minimum value in a list.
func Min[T constraints.Ordered](first T, rest ...T) T {
	min := first
	for _, val := range rest {
		if val < min {
			min = val
		}
	}
	return min
}

// Chunk splits the slice into chunks of size `size`. Returns a slice of slices.
// Chunk size must be greater than or equal to 1.
func Chunk[T any](slice []T, size int) [][]T {
	if size < 1 {
		panic("chunk size must be greater than or equal to 1")
	}

	chunks := make([][]T, 0, 1+(len(slice)-1)/size)
	for size < len(slice) {
		slice, chunks = slice[size:], append(chunks, slice[0:size:size])
	}
	chunks = append(chunks, slice)

	return chunks
}
