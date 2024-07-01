package codenav

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"slices"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/scip/bindings/go/scip"
	orderedmap "github.com/wk8/go-ordered-map/v2"
	"go.opentelemetry.io/otel/attribute"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/codeintel/core"
	"github.com/sourcegraph/sourcegraph/internal/collections"
	"github.com/sourcegraph/sourcegraph/internal/observation"
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
	trace observation.TraceLogger,
	repo types.Repo,
	commit api.CommitID,
	identifier string,
	language string,
) (orderedmap.OrderedMap[core.RepoRelPath, candidateFile], error) {
	if identifier == "" {
		return *orderedmap.New[core.RepoRelPath, candidateFile](), nil
	}
	var contextLines int32 = 0
	patternType := "standard"
	repoName := fmt.Sprintf("^%s$", repo.Name)
	resultMap := *orderedmap.New[core.RepoRelPath, candidateFile]()
	// TODO: This should be dependent on the number of requested usages, with a configured global limit
	countLimit := 1000
	searchQuery := fmt.Sprintf("type:file repo:%s rev:%s language:%s count:%d case:yes /\\b%s\\b/", repoName, string(commit), language, countLimit, identifier)

	trace.Info("Running query", log.String("q", searchQuery))

	plan, err := client.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, search.Streaming, &contextLines)
	if err != nil {
		return resultMap, err
	}
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return resultMap, err
	}
	nonFileMatches := 0
	inconsistentFilepaths := 0
	duplicatedFilepaths := collections.NewSet[string]()
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
					trace.Warn("Failed to create scip range from match range",
						log.String("error", err.Error()),
						log.String("matchRange", fmt.Sprintf("%+v", matchRange)),
					)
					continue
				}
				matchCount += 1
				matches = append(matches, scipRange)
			}
		}
		// OK to use Unchecked method here as search API only returns repo-root relative paths
		_, alreadyPresent := resultMap.Set(core.NewRepoRelPathUnchecked(path), candidateFile{
			matches:             scip.SortRanges(matches),
			didSearchEntireFile: !fileMatch.LimitHit,
		})
		if alreadyPresent {
			duplicatedFilepaths.Add(path)
		}
	}
	trace.AddEvent("findCandidateOccurrencesViaSearch", attribute.Int("matchCount", matchCount))

	if !duplicatedFilepaths.IsEmpty() {
		trace.Warn("Saw the duplicate file paths in search results", log.String("paths", duplicatedFilepaths.String()))
	}
	if nonFileMatches != 0 {
		trace.Warn("Saw non file match in search results. The `type:file` on the query should guarantee this")
	}
	if inconsistentFilepaths != 0 {
		trace.Warn("Saw mismatched file paths between chunk matches in the same FileMatch. Report this to the search-platform")
	}

	return resultMap, nil
}

type symbolData struct {
	range_ scip.Range
	kind   string
}

func (s *symbolData) Range() scip.Range {
	return s.range_
}

// symbolSearchResult maps file paths to a list of symbols sorted by range
type symbolSearchResult struct {
	inner orderedmap.OrderedMap[core.RepoRelPath, []symbolData]
}

func (s *symbolSearchResult) Contains(path core.RepoRelPath, range_ scip.Range) bool {
	if symbols, ok := s.inner.Get(path); ok {
		_, found := slices.BinarySearchFunc(symbols, range_, func(s1 symbolData, s2 scip.Range) int {
			return s1.range_.CompareStrict(s2)
		})
		return found
	}
	return false
}

func symbolSearch(
	ctx context.Context,
	client searchclient.SearchClient,
	trace observation.TraceLogger,
	repo types.Repo,
	identifier string,
	language string,
) (symbolSearchResult, error) {
	if identifier == "" {
		return symbolSearchResult{}, nil
	}
	var contextLines int32 = 0
	patternType := "standard"
	repoName := fmt.Sprintf("^%s$", repo.Name)
	// Using the same limit as the current web app
	countLimit := 50
	symbolQuery := fmt.Sprintf("type:symbol repo:%s lang:%s count:%d case:yes /\\b%s\\b/", repoName, language, countLimit, identifier)

	plan, err := client.Plan(ctx, "V3", &patternType, symbolQuery, search.Precise, search.Streaming, &contextLines)
	if err != nil {
		return symbolSearchResult{}, err
	}
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return symbolSearchResult{}, err
	}
	matchCount := 0

	resultMap := *orderedmap.New[core.RepoRelPath, []symbolData]()
	for _, streamResult := range stream.Results {
		fileMatch, ok := streamResult.(*result.FileMatch)
		if !ok {
			continue
		}
		symbolDatas := genslices.MapFilter(fileMatch.Symbols, func(symbol *result.SymbolMatch) (symbolData, bool) {
			scipRange, err := scip.NewRange([]int32{
				int32(symbol.Symbol.Range().Start.Line),
				int32(symbol.Symbol.Range().Start.Character),
				int32(symbol.Symbol.Range().End.Line),
				int32(symbol.Symbol.Range().End.Character),
			})
			if err != nil {
				return symbolData{}, false
			}
			return symbolData{
				range_: scipRange,
				kind:   symbol.Symbol.Kind,
			}, true
		})
		slices.SortFunc(symbolDatas, func(s1 symbolData, s2 symbolData) int {
			return s1.range_.CompareStrict(s2.range_)
		})
		matchCount += len(symbolDatas)
		resultMap.Set(core.NewRepoRelPathUnchecked(fileMatch.Path), symbolDatas)
	}
	trace.AddEvent("symbolSearch", attribute.Int("matchCount", matchCount))

	return symbolSearchResult{resultMap}, nil
}

func nameFromGlobalSymbol(symbol *scip.Symbol) (string, bool) {
	if len(symbol.Descriptors) == 0 || symbol.Descriptors[0].Suffix == scip.Descriptor_Local {
		return "", false
	}
	return symbol.Descriptors[len(symbol.Descriptors)-1].Name, true
}

// sliceRangeFromReader returns the substring corresponding to the given single-line range.
// It fails if the range spans multiple lines or it is out-of-bounds for the reader
func sliceRangeFromReader(reader io.Reader, range_ scip.Range) (substr string, err error) {
	if range_.Start.Line != range_.End.Line {
		return "", errors.New("symbol range spans multiple lines")
	}

	scanner := bufio.NewScanner(reader)
	for i := int32(0); scanner.Scan() && i <= range_.Start.Line; i++ {
		if i == range_.Start.Line {
			line := scanner.Text()
			if len(line) < int(range_.End.Character) {
				return "", errors.New("symbol range is out-of-bounds")
			}
			// FIXME(issue: GRAPH-715): wrong (less wrong would be to use rune offsets, actually correct needs encoding of the string _and_ the scip.Range)
			return line[range_.Start.Character:range_.End.Character], nil
		}
	}
	return "", errors.New("symbol range is out-of-bounds")
}
