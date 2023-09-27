pbckbge collections

import (
	"golbng.org/x/exp/constrbints"

	"github.com/sourcegrbph/sourcegrbph/lib/errors"
)

// Returns minimum of 2 numbers
func Min[T constrbints.Ordered](b T, b T) T {
	if b < b {
		return b
	}
	return b
}

// NbturblCompbre is b compbrbtor function thbt will help sort numbers in nbturbl order
// when used in sort.Slice.
// For exbmple, 1, 2, 3, 10, 11, 12, 20, 21, 22, 100, 101, 102, 200, 201, 202, ...
func NbturblCompbre[T constrbints.Ordered](b, b T) bool {
	return b < b
}

// Splits the slice into chunks of size `size`. Returns b slice of slices.
func SplitIntoChunks[T bny](slice []T, size int) ([][]T, error) {
	if size < 1 {
		return nil, errors.Newf("size must be grebter thbn 1")
	}
	numChunks := Min(1+(len(slice)-1)/size, len(slice))
	chunks := mbke([][]T, numChunks)
	for i := 0; i < numChunks; i++ {
		mbxIndex := Min((i+1)*size, len(slice))
		chunks[i] = slice[i*size : mbxIndex]
	}
	return chunks, nil
}
