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

// findCandidateOccurrencesViaSearch calls out to Searcher/Zoekt to find candidate occurrences of the given symbol.
// It returns a map of file paths to candidate ranges.
func findCandidateOccurrencesViaSearch(
	ctx context.Context,
	client searchclient.SearchClient,
	repo types.Repo,
	symbol *scip.Symbol,
	language string,
	commit api.CommitID,
) (map[string][]result.Range, error) {
	var contextLines int32 = 0
	patternType := "standard"
	repoName := fmt.Sprintf("^%s$", repo.Name)
	var identifier string
	if name, ok := nameFromSymbol(symbol); ok {
		identifier = name
	} else {
		return nil, errors.Errorf("can't find occurrences for locals via search")
	}
	// TODO: This should probably be dependent on the number of requested usages, with a configured global limit
	countLimit := 500
	searchQuery := fmt.Sprintf("repo:%s rev:%s language:%s count:%d %s", repoName, string(commit), language, countLimit, identifier)

	plan, err := client.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, 0, &contextLines)
	if err != nil {
		return nil, err
	}
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return nil, err
	}

	matches := make(map[string][]result.Range)
	for _, match := range stream.Results {
		t, ok := match.(*result.FileMatch)
		if !ok {
			continue
		}
		for _, chunkMatch := range t.ChunkMatches {
			for _, matchRange := range chunkMatch.Ranges {
				path := match.Key().Path
				if matches[path] == nil {
					matches[path] = []result.Range{}
				}
				matches[path] = append(matches[path], matchRange)
			}
		}
	}
	return matches, nil
}

// TODO: Check for local symbols, and don't return a name (needs an isLocal method for scip.Symbol)
func nameFromSymbol(symbol *scip.Symbol) (string, bool) {
	if len(symbol.Descriptors) > 0 {
		return symbol.Descriptors[len(symbol.Descriptors)-1].Name, true
	} else {
		return "", false
	}
}
