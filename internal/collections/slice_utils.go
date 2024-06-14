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
	End   int // End is exclusive, End >= Start
}

func (r HalfOpenRange) IsEmpty() bool {
	return r.Start == r.End
}

func (r HalfOpenRange) Len() int {
	return r.End - r.Start
}

// BinarySearchRangeFunc does a binary search over the underlying
// sorted slice x for the target value, and returns the range of
// indexes which equals the target value.
//
// The length of the returned range indicates the number of matched values.
// So if there was no match, the returned range will be empty.
// In that case, the Start value can be used for insertion.
func BinarySearchRangeFunc[S ~[]E, E, T any](x S, target T, cmp func(E, T) int) HalfOpenRange {
	insertIndex, found := slices.BinarySearchFunc(x, target, cmp)
	if !found {
		return HalfOpenRange{insertIndex, insertIndex}
	}
	start := insertIndex
	if subRange := BinarySearchRangeFunc(x[:start], target, cmp); !subRange.IsEmpty() {
		start = subRange.Start
	}
	end := insertIndex + 1
	if subRange := BinarySearchRangeFunc(x[end:], target, cmp); !subRange.IsEmpty() {
		// subRange.End will have been computed from the start of the sub-slice
		end = end + subRange.End
	}
	return HalfOpenRange{Start: start, End: end}
}
