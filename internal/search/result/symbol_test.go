package result

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/sourcegraph/go-lsp"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/stretchr/testify/require"
)

func TestSymbolRange(t *testing.T) {
	t.Run("unescaped pattern", func(t *testing.T) {
		want := lsp.Range{
			Start: lsp.Position{Line: 0, Character: 37},
			End:   lsp.Position{Line: 0, Character: 40},
		}
		got := Symbol{Line: 1, Name: "baz", Pattern: `/^bar() { var regex = \/.*\\\/\/; function baz() { }  } $/`}.Range()
		if diff := cmp.Diff(want, got); diff != "" {
			t.Fatal(diff)
		}
	})
}

func TestSymbolURL(t *testing.T) {
	repoA := types.RepoName{Name: "repo/A", ID: 1}
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
					Name:    "testsymbol",
					Line:    3,
					Pattern: `/^abc testsymbol def$/`,
				},
			},
			url: "/repo/A/-/blob/A#L3:5-3:15",
		},
		{
			name: "with rev",
			symbol: SymbolMatch{
				File: &fileAB,
				Symbol: Symbol{
					Name:    "testsymbol",
					Line:    3,
					Pattern: `/^abc testsymbol def$/`,
				},
			},
			url: "/repo/A@testrev/-/blob/B#L3:5-3:15",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			u := tc.symbol.URL().String()
			require.Equal(t, tc.url, u)
		})
	}
}
