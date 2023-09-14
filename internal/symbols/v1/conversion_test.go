package v1

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_Search_SymbolsResponse_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original search.SymbolsResponse) bool {
		if !symbolsResponseWithinInt32(original) {
			return true // skip
		}

		var originalProto SearchResponse
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsResponse diff (-want +got):\n%s", diff)
	}
}

func Test_Search_SymbolsParameters_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original search.SymbolsParameters) bool {
		if !symbolsParametersWithinInt32(original) {
			return true // skip
		}

		var originalProto SearchRequest
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsParameters diff (-want +got):\n%s", diff)
	}
}

func Test_Result_Symbol_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original result.Symbol) bool {
		if !symbolWithinInt32(original) {
			return true // skip
		}

		var originalProto SearchResponse_Symbol
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_SymbolInfo_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original *types.SymbolInfo) bool {
		if original != nil {
			defRange := original.Definition.Range
			if defRange != nil && !rangeWithinInt32(*defRange) {
				return true // skip
			}
		}

		var originalProto SymbolInfoResponse
		originalProto.FromInternal(original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolInfo diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_SymbolInfo_ProtoRoundTripNil(t *testing.T) {
	// Make sure a nil SymbolInfo is returned as nil.
	var originalProto SymbolInfoResponse
	originalProto.FromInternal(nil)
	converted := originalProto.ToInternal()

	var expect *types.SymbolInfo
	if diff := cmp.Diff(expect, converted, cmpopts.EquateEmpty()); diff != "" {
		t.Errorf("SymbolInfo diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_LocalCodeIntelPayload_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original types.LocalCodeIntelPayload) bool {
		if !localCodeIntelPayloadWithinInt32(original) {
			return true // skip
		}

		var originalProto LocalCodeIntelResponse
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(&original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("LocalCodeIntelPayload diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_Symbol_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original types.Symbol) bool {
		if !localCodeIntelSymbolWithinInt32(original) {
			return true // skip
		}

		var originalProto LocalCodeIntelResponse_Symbol
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted, cmpopts.EquateEmpty()); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_RepoCommitPath_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original types.RepoCommitPath) bool {

		var originalProto RepoCommitPath
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("RepoCommitPath diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_Range_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original types.Range) bool {
		if !rangeWithinInt32(original) {
			return true // skip
		}

		var originalProto Range
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Range diff (-want +got):\n%s", diff)
	}
}

func Test_Internal_Types_Point_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original types.Point) bool {
		if !pointWithinInt32(original) {
			return true // skip
		}

		var originalProto Point
		originalProto.FromInternal(&original)

		converted := originalProto.ToInternal()

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Point diff (-want +got):\n%s", diff)
	}
}

// These helper functions help ensure that testing/quick doesn't generate
// int values that are outside the range of the int32 types in the protobuf definitions.

// In our application code, these values shouldn't be outside the range of int32:
//
//   - symbol.Line / Character: 2^31-1 lines / line length is highly unlikely to be exceeded in a real codebase
//   - symbolsParameters.Timeout: 2^31 - 1 is ~68 years, which nobody will ever set
//   - symbolsParameters.First: Assuming that each symbol is at least three characters long, 2^31 symbols is would be
//     a ~17 gigabyte file, which is unlikely to be exceeded in a real codebase
func symbolsResponseWithinInt32(r search.SymbolsResponse) bool {
	for _, s := range r.Symbols {
		if !withinInt32(s.Line, s.Character) {
			return false
		}
	}

	return true
}

func localCodeIntelPayloadWithinInt32(p types.LocalCodeIntelPayload) bool {
	for _, s := range p.Symbols {
		if !localCodeIntelSymbolWithinInt32(s) {
			return false
		}
	}

	return true
}

func localCodeIntelSymbolWithinInt32(s types.Symbol) bool {
	ranges := []types.Range{s.Def}
	ranges = append(ranges, s.Refs...)

	for _, r := range ranges {
		if !rangeWithinInt32(r) {
			return false
		}
	}

	return true
}
func pointWithinInt32(p types.Point) bool {
	return withinInt32(p.Row, p.Column)
}

func symbolsParametersWithinInt32(s search.SymbolsParameters) bool {
	return withinInt32(s.First)
}

func rangeWithinInt32(r types.Range) bool {
	return withinInt32(r.Row, r.Column, r.Length)
}

// Normally, our line/char fields should be within the range of int32 anyway (2^31-1)
func symbolWithinInt32(s result.Symbol) bool {
	return withinInt32(s.Line, s.Character)
}

func withinInt32(xs ...int) bool {
	for _, x := range xs {
		if x < math.MinInt32 || x > math.MaxInt32 {
			return false
		}
	}

	return true
}
