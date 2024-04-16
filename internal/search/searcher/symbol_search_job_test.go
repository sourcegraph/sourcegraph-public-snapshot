package searcher

import (
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

func Test_symbolsToMatches(t *testing.T) {
	type fileType struct {
		Path     string
		Language string
		Symbols  []string
	}

	fixture := []fileType{
		{Path: "path1", Language: "Go", Symbols: []string{"sym1"}},
		{Path: "path2", Language: "YAML", Symbols: []string{"sym1", "sym2"}},
	}

	input := []result.Symbol{}
	for _, file := range fixture {
		for _, symbol := range file.Symbols {
			input = append(input, result.Symbol{Path: file.Path, Name: symbol, Language: file.Language})
		}
	}

	output := symbolsToMatches(input, types.MinimalRepo{Name: "somerepo"}, "abcdef", "abcdef")

	got := []fileType{}
	for _, match := range output {
		fileMatch := match.(*result.FileMatch)
		symbols := []string{}
		for _, symbol := range fileMatch.Symbols {
			symbols = append(symbols, symbol.Symbol.Name)
		}
		got = append(got, fileType{
			Path:     fileMatch.Path,
			Language: fileMatch.PreciseLanguage,
			Symbols:  symbols,
		})
	}

	want := fixture

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("symbolsToMatches() returned diff (-got +want):\n%s", diff)
	}
}
