package collections

import (
	"golang.org/x/exp/constraints"
	"slices"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

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
	numChunks := min(1+(len(slice)-1)/size, len(slice))
	chunks := make([][]T, numChunks)
	for i := range numChunks {
		maxIndex := min((i+1)*size, len(slice))
		chunks[i] = slice[i*size : maxIndex]
	}
	return chunks, nil
}

type HalfOpenRange struct {
	Start int // Start is inclusive
	End   int // End is exclusive
}

func BinarySearchRangeFunc[S ~[]E, E, T any](x S, target T, cmp func(E, T) int) (HalfOpenRange, bool) {
	insertIndex, found := slices.BinarySearchFunc(x, target, cmp)
	if !found {
		return HalfOpenRange{insertIndex, insertIndex + 1}, false
	}
	start := insertIndex
	for ; start >= 0 && cmp(x[start], target) == 0; start-- {
	}
	start += 1
	end := insertIndex
	for ; end < len(x) && cmp(x[end], target) == 0; end++ {
	}
	return HalfOpenRange{Start: start, End: end}, true
}
