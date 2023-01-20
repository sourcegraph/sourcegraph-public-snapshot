package types

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_LocalCodeIntelPayload_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original LocalCodeIntelPayload) bool {
		if !localCodeIntelPayloadWithinInt32(original) {
			return true // skip this iteration
		}

		var converted LocalCodeIntelPayload
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LocalCodeIntelPayload diff (-want +got):\n%s", diff)
	}
}

func Test_Point_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original Point) bool {
		if !pointWithinInt32(original) {
			return true // skip this iteration
		}

		var converted Point
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Point diff (-want +got):\n%s", diff)
	}
}

func Test_Range_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original Range) bool {
		if !rangeWithinInt32(original) {
			return true // skip this iteration
		}

		var converted Range
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Range diff (-want +got):\n%s", diff)
	}
}

func Test_RepoCommitPath_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original RepoCommitPath) bool {
		var converted RepoCommitPath
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RepoCommitPath diff (-want +got):\n%s", diff)
	}
}

func Test_SymbolInfo_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original SymbolInfo) bool {
		defRange := original.Definition.Range
		if defRange != nil && !rangeWithinInt32(*defRange) {
			return true // skip
		}

		var converted SymbolInfo
		converted.FromProto(original.ToProto())

		if diff := cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolInfo diff (-want +got):\n%s", diff)
	}

}

func Test_Symbol_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original Symbol) bool {
		if !symbolWithinInt32(original) {
			return true // skip this iteration
		}

		var converted Symbol
		converted.FromProto(original.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-want +got):\n%s", diff)
	}
}

// These withinInt32 functions are necessary to ensure that
// quick.Check doesn't generate int values that are larger than the
// specified int32 types in the protobuf definitions. (Otherwise, we'll get
// spurious diffs due to the loss of precision when converting to/from int <-> int32.)
//
// Our actual application code doesn't need to worry about this because we
// won't actually encounter files / lines that are 2^31 lines long/wide, etc.
func localCodeIntelPayloadWithinInt32(p LocalCodeIntelPayload) bool {
	for _, s := range p.Symbols {
		if !symbolWithinInt32(s) {
			return false
		}
	}

	return true
}

func symbolWithinInt32(s Symbol) bool {
	ranges := []Range{s.Def}
	ranges = append(ranges, s.Refs...)

	for _, r := range ranges {
		if !rangeWithinInt32(r) {
			return false
		}
	}

	return true
}

func rangeWithinInt32(r Range) bool {
	return withinInt32(r.Row, r.Column, r.Length)
}

func pointWithinInt32(p Point) bool {
	return withinInt32(p.Row, p.Column)
}

func withinInt32(xs ...int) bool {
	for _, x := range xs {
		if x < math.MinInt32 || x > math.MaxInt32 {
			return false
		}
	}

	return true

}
