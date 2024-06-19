package codenav

import (
	"context"
	"fmt"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	"github.com/sourcegraph/sourcegraph/internal/observation"
	orderedmap "github.com/wk8/go-ordered-map/v2"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchclient "github.com/sourcegraph/sourcegraph/internal/search/client"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type candidateFile struct {
	matches             []scip.Range // Guaranteed to be sorted
	didSearchEntireFile bool         // Or did we hit the search count limit?
}

// findCandidateOccurrencesViaSearch calls out to Searcher/Zoekt to find candidate occurrences of the given symbol.
// It returns a map of file paths to candidate ranges.
func findCandidateOccurrencesViaSearch(
	ctx context.Context,
	client searchclient.SearchClient,
	logger log.Logger,
	trace observation.TraceLogger,
	repo types.Repo,
	commit api.CommitID,
	symbol *scip.Symbol,
	language string,
) (orderedmap.OrderedMap[string, candidateFile], int, error) {
	var contextLines int32 = 0
	patternType := "standard"
	repoName := fmt.Sprintf("^%s$", repo.Name)
	var identifier string
	resultMap := *orderedmap.New[string, candidateFile]()
	if name, ok := nameFromSymbol(symbol); ok {
		identifier = name
	} else {
		return resultMap, 0, errors.Errorf("can't find occurrences for locals via search")
	}
	// TODO: This should be dependent on the number of requested usages, with a configured global limit
	countLimit := 1000
	searchQuery := fmt.Sprintf("type:file repo:%s rev:%s language:%s count:%d case:yes /\\b%s\\b/", repoName, string(commit), language, countLimit, identifier)

	trace.Info("Running query", log.String("q", searchQuery))

	plan, err := client.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, search.Streaming, &contextLines)
	if err != nil {
		return resultMap, 0, err
	}
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return resultMap, 0, err
	}

	nonFileMatches := 0
	inconsistentFilepaths := 0
	matchCount := 0

	for _, streamResult := range stream.Results {
		fileMatch, ok := streamResult.(*result.FileMatch)
		if !ok {
			nonFileMatches += 1
			continue
		}
		path := fileMatch.Path
		matches := []scip.Range{}
		for _, chunkMatch := range fileMatch.ChunkMatches {
			for _, matchRange := range chunkMatch.Ranges {
				if path != streamResult.Key().Path {
					inconsistentFilepaths = 1
					continue
				}
				scipRange, err := scip.NewRange([]int32{
					int32(matchRange.Start.Line),
					int32(matchRange.Start.Column),
					int32(matchRange.End.Line),
					int32(matchRange.End.Column),
				})
				if err != nil {
					logger.Error("Failed to create scip range from match range",
						log.String("error", err.Error()),
						log.String("matchRange", fmt.Sprintf("%+v", matchRange)),
					)
					continue
				}
				matchCount += 1
				matches = append(matches, scipRange)
			}
		}
		_, alreadyPresent := resultMap.Set(path, candidateFile{
			matches:             scip.SortRanges(matches),
			didSearchEntireFile: !fileMatch.LimitHit,
		})
		if alreadyPresent {
			logger.Error("Saw the same filepath twice in search results", log.String("path", path))
		}
		if nonFileMatches != 0 {
			logger.Error("Saw non file match in search results. The `type:file` on the query should guarantee this")
		}
		if inconsistentFilepaths != 0 {
			logger.Error("Saw mismatched file paths between chunk matches in the same FileMatch. Report this to the search-platform")
		}
	}
	return resultMap, matchCount, nil
}

func nameFromSymbol(symbol *scip.Symbol) (string, bool) {
	if len(symbol.Descriptors) == 0 || symbol.Descriptors[0].Suffix == scip.Descriptor_Local {
		return "", false
	}
	return symbol.Descriptors[len(symbol.Descriptors)-1].Name, true
}
