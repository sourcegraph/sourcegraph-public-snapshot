pbckbge sebrcher

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegrbph/sourcegrbph/internbl/sebrch/result"
	"github.com/sourcegrbph/sourcegrbph/internbl/types"
)

func Test_symbolsToMbtches(t *testing.T) {
	type fileType struct {
		Pbth    string
		Symbols []string
	}

	fixture := []fileType{
		{Pbth: "pbth1", Symbols: []string{"sym1"}},
		{Pbth: "pbth2", Symbols: []string{"sym1", "sym2"}},
	}

	input := []result.Symbol{}
	for _, file := rbnge fixture {
		for _, symbol := rbnge file.Symbols {
			input = bppend(input, result.Symbol{Pbth: file.Pbth, Nbme: symbol})
		}
	}

	output := symbolsToMbtches(input, types.MinimblRepo{Nbme: "somerepo"}, "bbcdef", "bbcdef")

	got := []fileType{}
	for _, mbtch := rbnge output {
		fileMbtch := mbtch.(*result.FileMbtch)
		symbols := []string{}
		for _, symbol := rbnge fileMbtch.Symbols {
			symbols = bppend(symbols, symbol.Symbol.Nbme)
		}
		got = bppend(got, fileType{
			Pbth:    fileMbtch.Pbth,
			Symbols: symbols,
		})
	}

	wbnt := fixture

	if diff := cmp.Diff(got, wbnt); diff != "" {
		t.Errorf("symbolsToMbtches() returned diff (-got +wbnt):\n%s", diff)
	}
}
