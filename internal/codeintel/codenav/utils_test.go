// This file does not contain tests for utils.go, instead it contains utils
// for setting up the test environment for our codenav tests.
package codenav

import (
	"context"
	"strings"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/search"
	searchClient "github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
)

func scipToResultPosition(p scip.Position) result.Location {
	return result.Location{
		Line:   int(p.Line),
		Column: int(p.Character),
	}
}

func scipToResultRange(r scip.Range) result.Range {
	return result.Range{
		Start: scipToResultPosition(r.Start),
		End:   scipToResultPosition(r.End),
	}
}

// scipToSymbolMatch "reverse engineers" the lsp.Range function on result.Symbol
func scipToSymbolMatch(r scip.Range) *result.SymbolMatch {
	return &result.SymbolMatch{
		Symbol: result.Symbol{
			Line:      int(r.Start.Line + 1),
			Character: int(r.Start.Character),
			Name:      strings.Repeat("a", int(r.End.Character-r.Start.Character)),
		}}
}

type FakeSearchBuilder struct {
	fileMatches   []result.Match
	symbolMatches []result.Match
}

func FakeSearchClient() FakeSearchBuilder {
	return FakeSearchBuilder{
		fileMatches:   []result.Match{},
		symbolMatches: make([]result.Match, 0),
	}
}

func ChunkMatchWithLine(range_ scip.Range, line string) result.ChunkMatch {
	return result.ChunkMatch{
		Ranges:  []result.Range{scipToResultRange(range_)},
		Content: line,
		ContentStart: result.Location{
			Line:   int(range_.Start.Line),
			Column: 0,
		},
	}
}

func ChunkMatch(range_ scip.Range) result.ChunkMatch {
	return ChunkMatchWithLine(range_, "chonky")
}

func ChunkMatches(ranges ...scip.Range) []result.ChunkMatch {
	return genslices.Map(ranges, ChunkMatch)
}

func (b FakeSearchBuilder) WithFile(file string, matches ...result.ChunkMatch) FakeSearchBuilder {
	b.fileMatches = append(b.fileMatches, &result.FileMatch{
		File:         result.File{Path: file},
		ChunkMatches: matches,
	})
	return b
}

func (b FakeSearchBuilder) WithSymbols(file string, ranges ...scip.Range) FakeSearchBuilder {
	b.symbolMatches = append(b.symbolMatches, &result.FileMatch{
		File:    result.File{Path: file},
		Symbols: genslices.Map(ranges, scipToSymbolMatch),
	})
	return b
}

func (b FakeSearchBuilder) Build() searchClient.SearchClient {
	mockSearchClient := searchClient.NewMockSearchClient()
	mockSearchClient.PlanFunc.SetDefaultHook(func(_ context.Context, _ string, _ *string, query string, _ search.Mode, _ search.Protocol, _ *int32) (*search.Inputs, error) {
		return &search.Inputs{OriginalQuery: query}, nil
	})
	mockSearchClient.ExecuteFunc.SetDefaultHook(func(_ context.Context, s streaming.Sender, i *search.Inputs) (*search.Alert, error) {
		if strings.Contains(i.OriginalQuery, "type:file") {
			s.Send(streaming.SearchEvent{
				Results: b.fileMatches,
			})
		} else if strings.Contains(i.OriginalQuery, "type:symbol") {
			s.Send(streaming.SearchEvent{
				Results: b.symbolMatches,
			})
		}
		return nil, nil
	})
	return mockSearchClient
}
