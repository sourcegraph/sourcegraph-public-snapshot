package slices

import (
	"golang.org/x/exp/constraints"
)

// Min returns the minimum value in a list.
// It will panic if no arguments are provided.
func Min[T constraints.Ordered](list ...T) T {
	min := list[0]
	for _, val := range list {
		if val < min {
			min = val
		}
	}
	return min
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
