package types

import (
	"bytes"
	"encoding/binary"
	"io"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// EncodeRanges converts a sequence of integers representing a set of ranges within
// the a text document into a string of bytes as we store them in Postgres. Each range
// in the input must consist of four ordered components: start line, start character,
// end line, and end character. Multiple ranges can be represented by simply appending
// components.
//
// We make the assumption that the input ranges are ordered by their start line. When
// this is not the case, the encoding will still be correct but the delta encoding may
// not have as large of a savings.
func EncodeRanges(vs []int32) (buf []byte, _ error) {
	if len(vs) == 0 {
		return nil, nil
	} else if len(vs)%4 != 0 {
		return nil, errors.Newf("unexpected range length - have %d but expected a multiple of 4", len(vs))
	}

	// Optimistic capacity; we append exactly one or two bytes for each element in the
	// given array. We assume that most of the delta-encoded values will be small, so
	// we try not to over-allocate here.
	buf = make([]byte, 0, len(vs))

	// The following outer loop causes the inner loop to execute twice. The first invocation
	// of the loop delta-encodes all of the line numbers as a contiguous sequence. The second
	// invocation delta-encodes the character offset numbers.
	//
	// The delta encoding for line numbers should be impactful as it forms a non-decreasing
	// sequence, and multiple references to the same variable within a document are likely to
	// present some degree of locality.
	//
	// Delta encoding for character numbers should be impactful as well as the average distance
	// between character numbers should be less than the average line length in the document. The
	// vastly common (soft) character maximums fall between 80-120, which fits into a single byte.

	for i := 0; i <= 1; i++ {
		for j := i; j < len(vs); j += 2 {
			if j < 2 {
				buf = binary.AppendVarint(buf, int64(vs[j]))
			} else {
				buf = binary.AppendVarint(buf, int64(vs[j]-vs[j-2]))
			}
		}
	}

	return buf, nil
}

// DecodeRanges decodes the output of `EncodeRanges`, transforming the result into a SCIP range slice.
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

// decodeRangesFromReader decodes the output of `EncodeRanges`. The given reader is assumed
// to be non-empty.
func decodeRangesFromReader(r io.ByteReader) ([]int32, error) {
	deltas, err := readVarints(r)
	if err != nil {
		return nil, err
	}

	n := len(deltas)
	h := n / 2

	if n%4 != 0 {
		return nil, errors.Newf("unexpected number of encoded deltas - have %d but expected a multiple of 4", n)
	}

	// The following loop decodes two delta-encoded sequences of integers at once. The
	// first half of `deltas` is a sequence of start/end line number pairs; the latter
	// half of `deltas` is a similar sequence of character number pairs.

	combined := make([]int32, 0, n)
	combined = append(
		combined,
		deltas[0], // first line
		deltas[h], // first char
	)

	for j := 1; j < h; j++ {
		combined = append(
			combined,
			combined[2*j-2]+deltas[0+j], // delta decode based on last line added
			combined[2*j-1]+deltas[h+j], // delta decode based on last char added
		)
	}

	return combined, nil
}

// readVarints reads a slice of packed variable-length integers.
func readVarints(r io.ByteReader) (vs []int32, _ error) {
	for {
		v, err := binary.ReadVarint(r)
		if err != nil {
			if err == io.EOF {
				break
			}

			return nil, err
		}

		vs = append(vs, int32(v))
	}

	return vs, nil
}
