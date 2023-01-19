package types

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_LocalCodeIntelPayload_ProtoConversion(t *testing.T) {
	t.Skip("TODO(ggilmore): Add function to check all symbols to verify that they are within int32 bounds.")

	var diff string

	f := func(original LocalCodeIntelPayload) bool {
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

func Test_Point_ProtoConversion(t *testing.T) {
	f := func(original *Point) bool {
		if original != nil && !(withinInt32(original.Row) && withinInt32(original.Column)) {
			return true
		}

		var converted *Point
		converted.FromProto(original.ToProto())

		return cmp.Equal(original, converted)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_Range_ProtoConversion(t *testing.T) {
	f := func(original Range) bool {
		if !withinInt32(original.Row, original.Column, original.Length) {
			return true
		}

		var converted Range
		converted.FromProto(original.ToProto())

		return cmp.Equal(original, converted)
	}

	if err := quick.Check(f, nil); err != nil {
		t.Error(err)
	}
}

func Test_RepoCommitPath_ProtoConversion(t *testing.T) {
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

func Test_SymbolInfo_ProtoConversion(t *testing.T) {
	var diff string

	f := func(original SymbolInfo) bool {

		defRange := original.Definition.Range
		if defRange != nil && !withinInt32(defRange.Row, defRange.Column, defRange.Length) {
			return true
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

func Test_Symbol_ProtoConversion(t *testing.T) {
	var diff string

	f := func(original Symbol) bool {

		// TODO@ggilmore: not amazing ergonomics here with multiple calls to
		// to withinInt32. Maybe we should also have a function that validates a Range as a whole?
		ranges := []Range{original.Def}
		ranges = append(ranges, original.Refs...)

		for _, r := range ranges {
			if !withinInt32(r.Row, r.Column, r.Length) {
				return true
			}
		}

		var converted Symbol
		converted.FromProto(original.ToProto())

		if diff := cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-want +got):\n%s", diff)
	}
}

func withinInt32(xs ...int) bool {
	for _, x := range xs {
		if x < math.MinInt32 || x > math.MaxInt32 {
			return false
		}
	}

	return true

}
