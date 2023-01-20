package search

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
)

func Test_SymbolsParameters_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original SymbolsParameters) bool {
		if !symbolsParametersWithinInt32(original) {
			return true // skip
		}

		var converted SymbolsParameters
		original.FromProto(converted.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsParameters diff (-want +got):\n%s", diff)
	}
}

func Test_SymbolsResponse_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original SymbolsResponse) bool {
		if !symbolsResponseWithinInt32(original) {
			return true // skip
		}

		var converted SymbolsResponse
		original.FromProto(converted.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("SymbolsResponse diff (-want +got):\n%s", diff)
	}
}

// These helper functions help ensure that testing/quick doesn't generate
// int values that are outside the range of the int32 types in the protobuf definitions.
//
// In our application code, these values shouldn't be outside the range of int32:
//
//   - symbol.Line / Character: 2^31-1 lines / line length is highly unlikely to be exceeded in a real codebase
//   - symbolsParameters.Timeout: 2^31 - 1 is ~68 years, which nobody will ever set
//   - symbolsParameters.First: Assuming that each symbol is at least three characters long, 2^31 symbols is would be
//     a ~17 gigabyte file, which is unlikely to be exceeded in a real codebase

func symbolsResponseWithinInt32(r SymbolsResponse) bool {
	for _, symbol := range r.Symbols {
		for _, number := range []int{symbol.Line, symbol.Character} {
			if number < math.MinInt32 || number > math.MaxInt32 {
				return false
			}
		}
	}

	return true
}

func symbolsParametersWithinInt32(s SymbolsParameters) bool {
	for _, number := range []int{s.Timeout, s.First} {
		if number < math.MinInt32 || number > math.MaxInt32 {
			return false
		}
	}

	return true
}
