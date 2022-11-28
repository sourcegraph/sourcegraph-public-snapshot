package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/scip/bindings/go/scip"
)

func TestRangeEncoding(t *testing.T) {
	ranges := []int32{
		// single-line
		100, 10, 100, 20,
		101, 15, 101, 25,
		103, 16, 103, 26,
		103, 31, 103, 41,
		103, 55, 103, 65,
		151, 10, 151, 20,
		152, 15, 152, 25,
		154, 25, 154, 35,
		154, 50, 154, 60,

		// multi-line
		200, 10, 205, 20,
		201, 15, 206, 25,
		203, 16, 208, 26,
		203, 31, 208, 41,
		203, 55, 208, 65,
		251, 10, 256, 20,
		252, 15, 257, 25,
		254, 25, 259, 35,
		254, 50, 259, 60,
	}

	encoded, err := EncodeRanges(ranges)
	if err != nil {
		t.Fatalf("unexpected error encoding ranges: %s", err)
	}

	// Internal decode
	decodedFlattenedRanges, err := DecodeFlattenedRanges(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding ranges: %s", err)
	}
	if diff := cmp.Diff(ranges, decodedFlattenedRanges); diff != "" {
		t.Fatalf("unexpected ranges (-want +got):\n%s", diff)
	}

	// External decode
	decodedSCIPRanges, err := DecodeRanges(encoded)
	if err != nil {
		t.Fatalf("unexpected error decoding ranges: %s", err)
	}
	expectedSCIPRanges := make([]*scip.Range, 0, len(ranges)/4)
	for i := 0; i < len(ranges); i += 4 {
		expectedSCIPRanges = append(expectedSCIPRanges, scip.NewRange(ranges[i:i+4]))
	}
	if diff := cmp.Diff(expectedSCIPRanges, decodedSCIPRanges); diff != "" {
		t.Fatalf("unexpected ranges (-want +got):\n%s", diff)
	}
}
