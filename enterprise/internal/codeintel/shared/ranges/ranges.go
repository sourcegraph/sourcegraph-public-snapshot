package ranges

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EncodeRanges converts a sequence of integers representing a set of ranges within the a text
// document into a string of bytes as we store them in Postgres. Each range in the input must
// consist of four ordered components: start line, start character, end line, and end character.
// Multiple ranges can be represented by simply appending components.
//
// We make the assumption that the input ranges are ordered by their start line. When this is not
// the case, the encoding will still be correct but the delta encoding may not have as large of a
// savings.
func EncodeRanges(values []int32) (buf []byte, _ error) {
	n := len(values)
	if n == 0 {
		return nil, nil
	} else if n%4 != 0 {
		return nil, errors.Newf("unexpected range length - have %d but expected a multiple of 4", n)
	}

	// Partition the given range quads into the `shuffled` slice. We de-interlace each component of the
	// ranges and "column-orient" each component (all start lines packed together, etc) and delta-encode
	// each of the quadrants.
	//
	// - Q1 stores delta-encoded start lines, which produces small integers.
	// - Q2 stores delta-encoded start characters, which produces runs of zeroes if occurrences happen at
	//   the same column. This is pretty common in generated code, or for common things that occur in the
	//   language syntax (receiver of a Go method, etc).
	// - Q3 stores delta-encoded start/end line distances, which should result in a long run of zeros as
	//   the start/end line/character distances should not generally change between occurrences.
	// - Q4 stores delta-encoded start/end character distances, which should result in a long run of zeros.

	var (
		q1Offset = n / 4 * 0
		q2Offset = n / 4 * 1
		q3Offset = n / 4 * 2
		q4Offset = n / 4 * 3
		shuffled = make([]int32, n)
	)

	for rangeIndex, rangeOffset := 0, 0; rangeOffset < n; rangeIndex, rangeOffset = rangeIndex+1, rangeOffset+4 {
		var (
			// extract current values
			startLine         = values[rangeOffset+0]
			startCharacter    = values[rangeOffset+1]
			lineDistance      = values[rangeOffset+2] - values[rangeOffset+0]
			characterDistance = values[rangeOffset+3] - values[rangeOffset+1]
		)

		var (
			// extract previous range values
			previousStartLine         int32
			previousStartCharacter    int32
			previousLineDistance      int32
			previousCharacterDistance int32
		)
		if rangeIndex != 0 {
			previousIndex := (rangeIndex - 1) * 4
			previousStartLine = values[previousIndex+0]
			previousStartCharacter = values[previousIndex+1]
			previousLineDistance = values[previousIndex+2] - values[previousIndex+0]
			previousCharacterDistance = values[previousIndex+3] - values[previousIndex+1]
		}

		// delta-encode and store into target location in array
		shuffled[q1Offset+rangeIndex] = startLine - previousStartLine
		shuffled[q2Offset+rangeIndex] = startCharacter - previousStartCharacter
		shuffled[q3Offset+rangeIndex] = lineDistance - previousLineDistance
		shuffled[q4Offset+rangeIndex] = characterDistance - previousCharacterDistance
	}

	// As Q3 and Q4 will likely have the forms:
	//
	// - `[q3 initial value], 0, 0, 0, ....`
	// - `[q4 initial value], 0, 0, 0, ....`
	//
	// We can reverse the values in Q4 so that the runs of zeros are contiguous. This will increase
	// the length of the run of zeros that can be run-length encoded.

	for i, j := q4Offset, n-1; i < j; i, j = i+1, j-1 {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	// Convert slice of ints into a packed byte slice. This will also run-length encode runs of zeros
	// which should be extremely common the second quarter of the shuffled array, as the vast majority
	// of occurrences will be single-lined.

	return writeVarints(shuffled), nil
}

// DecodeRanges decodes the output of `EncodeRanges`, transforming the result into a SCIP range
// slice.
func DecodeRanges(encoded []byte) ([]*scip.Range, error) {
	flattenedRanges, err := DecodeFlattenedRanges(encoded)
	if err != nil {
		return nil, err
	}

	n := len(flattenedRanges)
	ranges := make([]*scip.Range, 0, n/4)
	for i := 0; i < n; i += 4 {
		ranges = append(ranges, scip.NewRange(flattenedRanges[i:i+4]))
	}

	return ranges, nil
}

// DecodeFlattenedRanges decodes the output of `EncodeRanges`.
func DecodeFlattenedRanges(encoded []byte) ([]int32, error) {
	if len(encoded) == 0 {
		return nil, nil
	}

	return decodeRangesFromReader(bytes.NewReader(encoded))
}

// decodeRangesFromReader decodes the output of `EncodeRanges`.
func decodeRangesFromReader(r io.ByteReader) ([]int32, error) {
	values, err := readVarints(r)
	if err != nil {
		return nil, err
	}

	n := len(values)
	if n%4 != 0 {
		return nil, errors.Newf("unexpected number of encoded deltas - have %d but expected a multiple of 4", n)
	}

	var (
		q1Offset = n / 4 * 0
		q2Offset = n / 4 * 1
		q3Offset = n / 4 * 2
		q4Offset = n / 4 * 3
		combined = make([]int32, 0, n)
	)

	// Un-reverse Q4
	for i, j := q4Offset, n-1; i < j; i, j = i+1, j-1 {
		values[i], values[j] = values[j], values[i]
	}

	var (
		// Keep track of previous values for delta-decoding
		startLine         int32 = 0
		startCharacter    int32 = 0
		lineDistance      int32 = 0
		characterDistance int32 = 0
	)

	for i, j := 0, 0; j < n; i, j = i+1, j+4 {
		var (
			deltaEncodedStartLine         = values[q1Offset+i]
			deltaEncodedStartCharacter    = values[q2Offset+i]
			deltaEncodedLineDistance      = values[q3Offset+i]
			deltaEncodedCharacterDistance = values[q4Offset+i]
		)

		startLine += deltaEncodedStartLine
		startCharacter += deltaEncodedStartCharacter
		lineDistance += deltaEncodedLineDistance
		characterDistance += deltaEncodedCharacterDistance

		combined = append(
			combined,
			startLine,                        // start line
			startCharacter,                   // start character
			startLine+lineDistance,           // end line
			startCharacter+characterDistance, // end character
		)
	}

	return combined, nil
}

// writeVarints writes each of the given values as a varint into a by buffer. This function encodes
// runs of zeros as a single zero followed by the length of the run. The `readVarints` function will
// re-expand these runs of zeroes.
func writeVarints(values []int32) []byte {
	// Optimistic capacity; we append exactly one or two bytes for each non-zero element in the given
	// array. We assume that most of the values are small, so we try not to over-allocate here. We may
	// resize only once in the worst case.
	buf := make([]byte, 0, len(values))

	i := 0
	for i < len(values) {
		value := values[i]
		if value == 0 {
			runStart := i
			for i < len(values) && values[i] == 0 {
				i++
			}

			buf = binary.AppendVarint(buf, int64(0))
			buf = binary.AppendVarint(buf, int64(i-runStart))
			continue
		}

		buf = binary.AppendVarint(buf, int64(value))
		i++
	}

	return buf
}

// readVarints reads a sequence of varints from the given reader as encoded by `writeVarints`. When
// a zero-value is encountered, this function expects the next varint value to be the length of a run
// of zero values.
//
// The slice of values returned by this function will contain the expanded run of zeroes.
func readVarints(r io.ByteReader) (values []int32, _ error) {
	for {
		if value, ok, err := readVarint32(r); err != nil {
			return nil, err
		} else if !ok {
			break
		} else if value != 0 {
			// Regular value
			values = append(values, value)
			continue
		}

		// We read a zero value; read the length of the run and pad the output slice
		count, ok, err := readVarint32(r)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.New("expected length for run of zero values")
		}
		for ; count > 0; count-- {
			values = append(values, 0)
		}
	}

	return values, nil
}

// readVarint32 reads a single varint from the given reader. If the reader has no more content a
// false-valued flag is returned.
func readVarint32(r io.ByteReader) (int32, bool, error) {
	value, err := binary.ReadVarint(r)
	if err != nil {
		if err == io.EOF {
			return 0, false, nil
		}

		return 0, false, err
	}

	return int32(value), true, nil
}
