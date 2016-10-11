package langserver

import (
	"reflect"
	"sort"
	"testing"

	"sourcegraph.com/sourcegraph/sourcegraph/pkg/lsp"
)

func Test_resultSorter(t *testing.T) {
	type testcase struct {
		rawQuery   string
		allSymbols []lsp.SymbolInformation
		expResults []lsp.SymbolInformation
	}
	tests := []testcase{{
		rawQuery: "foo.bar",
		allSymbols: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "foo",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "asdf",
			Location: lsp.Location{URI: "foo.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "one", Name: "two",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
		expResults: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "foo",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "asdf",
			Location: lsp.Location{URI: "foo.go"},
			Kind:     lsp.SKFunction,
		}},
	}, {
		rawQuery: "foo bar",
		allSymbols: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "foo",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "asdf",
			Location: lsp.Location{URI: "foo.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "one", Name: "two",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
		expResults: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "foo",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}, {
			ContainerName: "asdf", Name: "asdf",
			Location: lsp.Location{URI: "foo.go"},
			Kind:     lsp.SKFunction,
		}},
	}, {
		// Just tests that 'is:exported' does not affect resultSorter
		// results, as filtering is done elsewhere in (*LangHandler).collectFromPkg
		rawQuery: "is:exported",
		allSymbols: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
		expResults: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
	}, {
		rawQuery: "",
		allSymbols: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
		expResults: []lsp.SymbolInformation{{
			ContainerName: "foo", Name: "bar",
			Location: lsp.Location{URI: "file.go"},
			Kind:     lsp.SKFunction,
		}},
	}}

	for _, test := range tests {
		results := resultSorter{query: parseQuery(test.rawQuery)}
		for _, s := range test.allSymbols {
			results.Collect(s)
		}
		sort.Sort(&results)
		if !reflect.DeepEqual(results.Results(), test.expResults) {
			t.Errorf("got %+v, but wanted %+v", results.Results(), test.expResults)
		}
	}
}
