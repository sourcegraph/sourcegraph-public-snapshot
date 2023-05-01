package collections

import (
	"golang.org/x/exp/constraints"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Returns minimum of 2 numbers
func Min[T constraints.Ordered](a T, b T) T {
	if a < b {
		return a
	}
	return b
}

// NaturalCompare is a comparator function that will help sort numbers in natural order
// when used in sort.Slice.
// For example, 1, 2, 3, 10, 11, 12, 20, 21, 22, 100, 101, 102, 200, 201, 202, ...
func NaturalCompare[T constraints.Ordered](a, b T) bool {
	return a < b
}

// Splits the slice into chunks of size `size`. Returns a slice of slices.
func SplitIntoChunks[T any](slice []T, size int) ([][]T, error) {
	if size < 1 {
		return nil, errors.Newf("size must be greater than 1")
	}
	numChunks := Min(1+(len(slice)-1)/size, len(slice))
	chunks := make([][]T, numChunks)
	for i := 0; i < numChunks; i++ {
		maxIndex := Min((i+1)*size, len(slice))
		chunks[i] = slice[i*size : maxIndex]
	}
	return chunks, nil
}
