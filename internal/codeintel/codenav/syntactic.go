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
)

func FindCandidateOccurrencesViaSearch(
	ctx context.Context,
	client searchclient.SearchClient,
	repo types.Repo,
	symbolName string,
	revision api.CommitID,
) (map[string][]result.Range, error) {
	var contextLines int32 = 0
	patternType := "standard"
	// TODO: Once we handle languages with ambiguous file extensions we'll need to be smarter about this.
	// Figure out why the enry call isn't working
	language := "java" // enry.GetLanguageByFilename(filepath.Base(args.Range.Path))
	repoName := fmt.Sprintf("^%s$", repo.Name)
	identifier := symbolName
	countLimit := 500
	searchQuery := fmt.Sprintf("repo:%s rev:%s language:%s count:%d %s", repoName, string(revision), language, countLimit, identifier)

	// fmt.Printf("Sending: query=%s\n", searchQuery)

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

func NameFromSymbol(symbol *scip.Symbol) string {
	lastDescriptor := ""
	if len(symbol.Descriptors) > 0 {
		lastDescriptor = symbol.Descriptors[len(symbol.Descriptors)-1].Name
	}
	return lastDescriptor
}

func ScoreMatch(symbol1 *scip.Symbol, symbol2 *scip.Symbol) float64 {
	return 9000.1
}

func RangesOverlap(zearcherRange result.Range, scipRange *scip.Range) bool {
	return zearcherRange.Start.Line == int(scipRange.Start.Line) &&
		zearcherRange.End.Line == int(scipRange.End.Line) &&
		zearcherRange.Start.Column <= int(scipRange.End.Character) &&
		int(scipRange.Start.Character) <= zearcherRange.End.Column
}
