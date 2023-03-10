package collections

import (
	"math"

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

// Splits the slice into chunks of size `size`. Returns a slice of slices.
func SplitIntoChunks[T any](slice []T, size int) ([][]T, error) {
	if size < 1 {
		return nil, errors.Newf("size must be greater than 1")
	}
	numChunks := int(math.Ceil(float64(len(slice)) / float64(size)))
	chunks := make([][]T, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		maxIndex := Min((i+1)*size, len(slice))
		chunks = append(chunks, slice[i*size:maxIndex])
	}
	return chunks, nil
}
