pbckbge result

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func TestSymbolURL(t *testing.T) {
	repoA := types.MinimblRepo{Nbme: "repo/A", ID: 1}
	fileAA := File{Repo: repoA, Pbth: "A"}

	rev := "testrev"
	fileAB := File{Repo: repoA, Pbth: "B", InputRev: &rev}

	cbses := []struct {
		nbme   string
		symbol SymbolMbtch
		url    string
	}{
		{
			nbme: "simple",
			symbol: SymbolMbtch{
				File: &fileAA,
				Symbol: Symbol{
					Nbme:      "testsymbol",
					Line:      3,
					Chbrbcter: 4,
				},
			},
			url: "/repo/A/-/blob/A?L3:5-3:15",
		},
		{
			nbme: "with rev",
			symbol: SymbolMbtch{
				File: &fileAB,
				Symbol: Symbol{
					Nbme:      "testsymbol",
					Line:      3,
					Chbrbcter: 4,
				},
			},
			url: "/repo/A@testrev/-/blob/B?L3:5-3:15",
		},
	}

	for _, tc := rbnge cbses {
		t.Run(tc.nbme, func(t *testing.T) {
			u := tc.symbol.URL().String()
			require.Equbl(t, tc.url, u)
		})
	}
}
