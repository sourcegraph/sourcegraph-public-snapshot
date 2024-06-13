package codenav

import (
	"context"
	"fmt"

	"github.com/sourcegraph/scip/bindings/go/scip"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchclient "github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type candidateFile struct {
	matches  []scip.Range // Guaranteed to be sorted
	complete bool         // Was this file searched in its entirety, or did we hit the search count limit?
}

// findCandidateOccurrencesViaSearch calls out to Searcher/Zoekt to find candidate occurrences of the given symbol.
// It returns a map of file paths to candidate ranges.
func findCandidateOccurrencesViaSearch(
	ctx context.Context,
	client searchclient.SearchClient,
	repo types.Repo,
	commit api.CommitID,
	symbol *scip.Symbol,
	language string,
) (map[string]candidateFile, error) {
	var contextLines int32 = 0
	patternType := "standard"
	repoName := fmt.Sprintf("^%s$", repo.Name)
	var identifier string
	if name, ok := nameFromSymbol(symbol); ok {
		identifier = name
	} else {
		return nil, errors.Errorf("can't find occurrences for locals via search")
	}
	// TODO: This should be dependent on the number of requested usages, with a configured global limit
	countLimit := 500
	searchQuery := fmt.Sprintf("type:file repo:%s rev:%s language:%s count:%d %s", repoName, string(commit), language, countLimit, identifier)

	plan, err := client.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, 0, &contextLines)
	if err != nil {
		return nil, err
	}
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return nil, err
	}

	results := make(map[string]candidateFile)
	for _, match := range stream.Results {
		fileMatch, ok := match.(*result.FileMatch)
		if !ok {
			// Is it worth asserting this can't happen?
			panic("non file match in search results. The `type:file` on the query should guarantee this")
			// continue
		}
		path := fileMatch.Path
		matches := []scip.Range{}
		for _, chunkMatch := range fileMatch.ChunkMatches {
			for _, matchRange := range chunkMatch.Ranges {
				if path != match.Key().Path {
					// Is it worth asserting this can't happen?
					panic("FileMatch with chunkMatch.Ranges from a different file")
				}
				scipRange := scip.NewRangeUnchecked([]int32{
					int32(matchRange.Start.Line),
					int32(matchRange.Start.Column),
					int32(matchRange.End.Line),
					int32(matchRange.End.Column),
				})
				matches = append(matches, scipRange)
			}
		}
		if len(matches) == 0 {
			// Is it worth asserting this can't happen?
			panic("FileMatch with no ranges")
		}
		results[path] = candidateFile{
			matches:  scip.SortRanges(matches),
			complete: !fileMatch.LimitHit,
		}
	}
	return results, nil
}

// TODO: Check for local symbols, and don't return a name (needs an isLocal method for scip.Symbol)
func nameFromSymbol(symbol *scip.Symbol) (string, bool) {
	if len(symbol.Descriptors) > 0 {
		return symbol.Descriptors[len(symbol.Descriptors)-1].Name, true
	} else {
		return "", false
	}
}
