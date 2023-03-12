package slices

import (
	"golang.org/x/exp/constraints"
)

// Min returns the minimum of two variables.
func Min[T constraints.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

// Chunk splits the slice into chunks of size `size`. Returns a slice of slices.
func Chunk[T any](slice []T, size int) [][]T {
	length := len(slice)
	numChunks := 1 + (length-1)/size
	chunks := make([][]T, numChunks)
	for i := 0; i < numChunks; i++ {
		chunks[i] = slice[i*size : Min((i+1)*size, length)]
	}

	return chunks
}
