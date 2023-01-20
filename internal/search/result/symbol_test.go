package result

import (
	"math"
	"testing"
	"testing/quick"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"

	"github.com/sourcegraph/sourcegraph/internal/types"
)

func TestSymbolURL(t *testing.T) {
	repoA := types.MinimalRepo{Name: "repo/A", ID: 1}
	fileAA := File{Repo: repoA, Path: "A"}

	rev := "testrev"
	fileAB := File{Repo: repoA, Path: "B", InputRev: &rev}

	cases := []struct {
		name   string
		symbol SymbolMatch
		url    string
	}{
		{
			name: "simple",
			symbol: SymbolMatch{
				File: &fileAA,
				Symbol: Symbol{
					Name:      "testsymbol",
					Line:      3,
					Character: 4,
				},
			},
			url: "/repo/A/-/blob/A?L3:5-3:15",
		},
		{
			name: "with rev",
			symbol: SymbolMatch{
				File: &fileAB,
				Symbol: Symbol{
					Name:      "testsymbol",
					Line:      3,
					Character: 4,
				},
			},
			url: "/repo/A@testrev/-/blob/B?L3:5-3:15",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.symbol.URL().String()
			require.Equal(t, tc.url, u)
		})
	}
}

func Test_Symbol_ProtoRoundTrip(t *testing.T) {
	var diff string

	f := func(original Symbol) bool {
		if !symbolWithinInt32(original) {
			return true // skip
		}

		var converted Symbol
		original.FromProto(converted.ToProto())

		if diff = cmp.Diff(original, converted); diff != "" {
			return false
		}

		return true
	}

	if err := quick.Check(f, nil); err != nil {
		t.Errorf("Symbol diff (-want +got):\n%s", diff)
	}
}

// small helper function to ensure that the symbol's line/char fields are within the range of int32
// since that is what is defined in the protobuf spec for the symbol message
//
// Normally, our line/char fields should be within the range of int32 anyway (2^31-1)
func symbolWithinInt32(s Symbol) bool {
	return s.Line >= math.MinInt32 && s.Line <= math.MaxInt32 &&
		s.Character >= math.MinInt32 && s.Character <= math.MaxInt32
}
