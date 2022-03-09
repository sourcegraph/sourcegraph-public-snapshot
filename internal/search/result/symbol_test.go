package result

import (
	"testing"

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
