package types

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

	var (
		q1Offset = n / 4 * 0
		q2Offset = n / 4 * 1
		q3Offset = n / 4 * 2
		q4Offset = n / 4 * 3
		shuffled = make([]int32, n)
	)

	// Partition the given range quads into a the `shuffled` slice. We de-interlace each component of
	// the ranges and "column-orient" each component (all start lines packed together, etc).

	for rangeIndex, rangeOffset := 0, 0; rangeOffset < n; rangeIndex, rangeOffset = rangeIndex+1, rangeOffset+4 {
		var (
			startLine      = values[rangeOffset+0]
			startCharacter = values[rangeOffset+1]
			endLine        = values[rangeOffset+2]
			endCharacter   = values[rangeOffset+3]
		)

		deltaEncodedStartLine := startLine
		if rangeIndex != 0 {
			deltaEncodedStartLine = startLine - values[(rangeIndex-1)*4]
		}

		shuffled[q1Offset+rangeIndex] = deltaEncodedStartLine         // Q1: delta-encoded start lines
		shuffled[q2Offset+rangeIndex] = endLine - startLine           // Q2: start line/end line deltas
		shuffled[q3Offset+rangeIndex] = startCharacter                // Q3: start character
		shuffled[q4Offset+rangeIndex] = endCharacter - startCharacter // Q4: start character/end character deltas
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

// decodeRangesFromReader decodes the output of `EncodeRanges`. The given reader is assumed to be
// non-empty.
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
		q1Offset        = n / 4 * 0
		q2Offset        = n / 4 * 1
		q3Offset        = n / 4 * 2
		q4Offset        = n / 4 * 3
		startLine int32 = 0
		combined        = make([]int32, 0, n)
	)

	for i, j := 0, 0; j < n; i, j = i+1, j+4 {
		var (
			deltaEncodedStartLine = values[q1Offset+i]
			lineDelta             = values[q2Offset+i]
			startCharacter        = values[q3Offset+i]
			characterDelta        = values[q4Offset+i]
		)

		// delta-decode start line
		startLine += deltaEncodedStartLine

		combined = append(
			combined,
			startLine,                     // start line
			startCharacter,                // start character
			startLine+lineDelta,           // end line
			startCharacter+characterDelta, // end character
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
