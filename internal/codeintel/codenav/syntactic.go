package codenav

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"slices"
	"strings"

	genslices "github.com/life4/genesis/slices"
	"github.com/sourcegraph/conc"
	conciter "github.com/sourcegraph/conc/iter"
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
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SYNTACTIC_USAGES_DOCUMENTS_CHUNK_SIZE is the batch size for SCIP documents and git diffs we load at a time.
//
// I collected traces for various sizes (on my local machine) and 20 gave me "nice looking" ones.
// In general I expect 100 documents to be on the "higher end" of the number of documents to retrieve
// for a single syntactic usage search and 5 concurrent queries and git requests seems like a reasonable
// trade-off for concurrency vs load.
const SYNTACTIC_USAGES_DOCUMENTS_CHUNK_SIZE = 20

type candidateMatch struct {
	range_             scip.Range
	surroundingContent string
}

type candidateFile struct {
	path    core.RepoRelPath
	matches []candidateMatch // Guaranteed to be sorted by range
}

type searchArgs struct {
	repo       api.RepoName
	commit     api.CommitID
	identifier string
	language   string
	countLimit int32
}

func lineForRange(match result.ChunkMatch, range_ result.Range) string {
	lines := strings.Split(match.Content, "\n")
	index := range_.Start.Line - match.ContentStart.Line
	if len(lines) <= index {
		return ""
	}
	return lines[index]
}

type candidateOccurrenceResult struct {
	candidateFiles []candidateFile
	limitHit       bool
}

// findCandidateOccurrencesViaSearch calls out to Searcher/Zoekt to find candidate occurrences of the given symbol.
// It returns a map of file paths to candidate ranges.
func findCandidateOccurrencesViaSearch(
	ctx context.Context,
	trace observation.TraceLogger,
	client searchclient.SearchClient,
	args searchArgs,
) (candidateOccurrenceResult, error) {
	if args.identifier == "" {
		return candidateOccurrenceResult{}, nil
	}
	resultMap := *orderedmap.New[core.RepoRelPath, candidateFile]()
	searchResults, err := executeQuery(ctx, client, trace, args, "file", 0)
	if err != nil {
		return candidateOccurrenceResult{}, err
	}

	nonFileMatches := 0
	inconsistentFilepaths := 0
	duplicatedFilepaths := collections.NewSet[string]()
	matchCount := 0
	limitHit := false
	for _, streamResult := range searchResults {
		fileMatch, ok := streamResult.(*result.FileMatch)
		if !ok {
			nonFileMatches += 1
			continue
		}
		// Pagination for search-based usages works at file granularity,
		// so don't consider partially searched files.
		if fileMatch.LimitHit {
			limitHit = true
			continue
		}
		path := fileMatch.Path
		matches := []candidateMatch{}
		for _, chunkMatch := range fileMatch.ChunkMatches {
			for _, matchRange := range chunkMatch.Ranges {
				if path != streamResult.Key().Path {
					inconsistentFilepaths = 1
					continue
				}
				scipRange, err := scipFromResultRange(matchRange)
				if err != nil {
					trace.Warn("Failed to create scip range from match range",
						log.String("error", err.Error()),
						log.String("matchRange", fmt.Sprintf("%+v", matchRange)),
					)
					continue
				}
				matchCount += 1
				matches = append(matches, candidateMatch{
					range_:             scipRange,
					surroundingContent: lineForRange(chunkMatch, matchRange),
				})
			}
		}
		slices.SortStableFunc(matches, func(s1, s2 candidateMatch) int {
			return s1.range_.CompareStrict(s2.range_)
		})
		// OK to use Unchecked method here as search API only returns repo-root relative paths
		relPath := core.NewRepoRelPathUnchecked(path)
		_, alreadyPresent := resultMap.Set(relPath, candidateFile{
			matches: matches,
			path:    relPath,
		})
		if alreadyPresent {
			duplicatedFilepaths.Add(path)
		}
	}
	trace.AddEvent("findCandidateOccurrencesViaSearch", attribute.Int("matchCount", matchCount))

	if !duplicatedFilepaths.IsEmpty() {
		trace.Warn("Saw duplicate file paths in search results", log.String("paths", duplicatedFilepaths.String()))
	}
	if nonFileMatches != 0 {
		trace.Warn("Saw non file match in search results. The `type:file` on the query should guarantee this")
	}
	if inconsistentFilepaths != 0 {
		trace.Warn("Saw mismatched file paths between chunk matches in the same FileMatch. Report this to the search-platform")
	}

	results := make([]candidateFile, 0, resultMap.Len())
	for pair := resultMap.Oldest(); pair != nil; pair = pair.Next() {
		results = append(results, pair.Value)
	}
	// FIXME: Edge case where limitHit is true but candidateFiles is empty => means we get more than limit matches in the same file.
	// This will cause the cursor to get stuck, as we can't make progress. Potential "fix" would be to return the incomplete
	// list of matches for that file (accepting that we'd be missing matches past the limit for that file).
	// This would end up adding the problematic file to the cursor, meaning it would be skipped on the next page.
	return candidateOccurrenceResult{
		candidateFiles: results,
		limitHit:       limitHit || matchCount >= int(args.countLimit),
	}, nil
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
	trace observation.TraceLogger,
	client searchclient.SearchClient,
	args searchArgs,
) (symbolSearchResult, error) {
	if args.identifier == "" {
		return symbolSearchResult{}, nil
	}
	// Using the same limit as the current web app
	searchResults, err := executeQuery(ctx, client, trace, args, "symbol", 0)
	if err != nil {
		return symbolSearchResult{}, err
	}

	matchCount := 0
	resultMap := *orderedmap.New[core.RepoRelPath, []symbolData]()
	for _, streamResult := range searchResults {
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

func buildQuery(args searchArgs, queryType string) string {
	repoName := fmt.Sprintf("^%s$", args.repo)
	wordBoundaryIdentifier := fmt.Sprintf("/\\b%s\\b/", args.identifier)
	return fmt.Sprintf(
		"case:yes type:%s repo:%s rev:%s language:%s count:%d %s",
		queryType, repoName, string(args.commit), args.language, args.countLimit, wordBoundaryIdentifier)
}

func executeQuery(
	ctx context.Context,
	client searchclient.SearchClient,
	trace observation.TraceLogger,
	args searchArgs,
	queryType string,
	surroundingLines int,
) (result.Matches, error) {
	searchQuery := buildQuery(args, queryType)
	patternType := "standard"
	contextLines := int32(surroundingLines)
	plan, err := client.Plan(ctx, "V3", &patternType, searchQuery, search.Precise, search.Streaming, &contextLines)
	if err != nil {
		return nil, err
	}
	trace.Info("Running query", log.String("query", searchQuery))
	stream := streaming.NewAggregatingStream()
	_, err = client.Execute(ctx, stream, plan)
	if err != nil {
		return nil, err
	}
	return stream.Results, nil
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

func scipFromResultRange(resultRange result.Range) (scip.Range, error) {
	return scip.NewRange([]int32{
		int32(resultRange.Start.Line),
		int32(resultRange.Start.Column),
		int32(resultRange.End.Line),
		int32(resultRange.End.Column),
	})
}

// symbolAtRange tries to look up the symbols at the given coordinates
// in a syntactic upload. If this function returns an error you should most likely
// log and handle it instead of rethrowing, as this could fail for a myriad of reasons
// (some broken invariant internally, network issue etc.)
// If this function doesn't error, the returned slice is guaranteed to be non-empty
func symbolAtRange(
	ctx context.Context,
	mappedIndex MappedIndex,
	args UsagesForSymbolArgs,
) (*scip.Symbol, error) {
	docOpt, err := mappedIndex.GetDocument(ctx, args.Path)
	if err != nil {
		return nil, err
	}
	doc, isSome := docOpt.Get()
	if !isSome {
		return nil, errors.New("no document found")
	}
	occs, err := doc.GetOccurrencesAtRange(ctx, args.SymbolRange)
	if err != nil {
		return nil, err
	}
	if len(occs) == 0 {
		return nil, errors.New("no occurrences found")
	}
	sym, err := scip.ParseSymbol(occs[0].Symbol)
	if err != nil {
		return nil, err
	}
	return sym, nil
}

func findSyntacticMatchesForCandidateFile(
	ctx context.Context,
	trace observation.TraceLogger,
	document MappedDocument,
	candidateFile candidateFile,
) ([]SyntacticMatch, []SearchBasedMatch) {
	filePath := candidateFile.path
	syntacticMatches := []SyntacticMatch{}
	searchBasedMatches := []SearchBasedMatch{}
	failedTranslationCount := 0
	for _, candidateMatch := range candidateFile.matches {
		foundSyntacticMatch := false
		occurrences, occErr := document.GetOccurrencesAtRange(ctx, candidateMatch.range_)
		if occErr != nil {
			failedTranslationCount += 1
			continue
		}
		for _, occ := range occurrences {
			if !scip.IsLocalSymbol(occ.Symbol) {
				foundSyntacticMatch = true
				syntacticMatches = append(syntacticMatches, SyntacticMatch{
					Path:               filePath,
					Range:              candidateMatch.range_,
					SurroundingContent: candidateMatch.surroundingContent,
					Symbol:             occ.Symbol,
					IsDefinition:       scip.SymbolRole_Definition.Matches(occ),
				})
			}
		}
		if !foundSyntacticMatch {
			searchBasedMatches = append(searchBasedMatches, SearchBasedMatch{
				Path:               filePath,
				Range:              candidateMatch.range_,
				SurroundingContent: candidateMatch.surroundingContent,
				IsDefinition:       false,
			})
		}
	}
	if failedTranslationCount != 0 {
		trace.Info("findSyntacticMatchesForCandidateFile", log.Int("failedTranslationCount", failedTranslationCount))
	}
	return syntacticMatches, searchBasedMatches
}

type fileMatches[T any] struct {
	path    core.RepoRelPath
	matches []T
}

func syntacticUsagesImpl(
	ctx context.Context,
	trace observation.TraceLogger,
	searchClient searchclient.SearchClient,
	mappedIndex MappedIndex,
	args UsagesForSymbolArgs,
) (SyntacticUsagesResult, PreviousSyntacticSearch, *SyntacticUsagesError) {
	searchSymbol, symErr := symbolAtRange(ctx, mappedIndex, args)
	if symErr != nil {
		return SyntacticUsagesResult{}, PreviousSyntacticSearch{}, &SyntacticUsagesError{
			Code:            SU_NoSymbolAtRequestedRange,
			UnderlyingError: symErr,
		}
	}
	language, langErr := languageFromFilepath(trace, args.Path)
	if langErr != nil {
		return SyntacticUsagesResult{}, PreviousSyntacticSearch{}, &SyntacticUsagesError{
			Code:            SU_FailedToSearch,
			UnderlyingError: langErr,
		}
	}
	symbolName, ok := nameFromGlobalSymbol(searchSymbol)
	if !ok {
		return SyntacticUsagesResult{}, PreviousSyntacticSearch{}, &SyntacticUsagesError{
			Code:            SU_FailedToSearch,
			UnderlyingError: errors.New("can't find syntactic occurrences for locals via search"),
		}
	}
	searchCoords := searchArgs{
		repo:       args.Repo.Name,
		commit:     args.Commit,
		identifier: symbolName,
		language:   language,
		// TODO: Assumes at least every third match is a search-based one
		countLimit: args.Limit * 3,
	}
	searchResult, searchErr := findCandidateOccurrencesViaSearch(ctx, trace, searchClient, searchCoords)
	if searchErr != nil {
		return SyntacticUsagesResult{}, PreviousSyntacticSearch{}, &SyntacticUsagesError{
			Code:            SU_FailedToSearch,
			UnderlyingError: searchErr,
		}
	}

	tasks, _ := genslices.ChunkEvery(searchResult.candidateFiles, SYNTACTIC_USAGES_DOCUMENTS_CHUNK_SIZE)
	results, err := conciter.MapErr(tasks, func(files *[]candidateFile) ([]fileMatches[SyntacticMatch], error) {
		paths := genslices.Map(*files, func(f candidateFile) core.RepoRelPath {
			return f.path
		})
		documents, err := mappedIndex.GetDocuments(ctx, paths)
		if err != nil {
			return nil, err
		}
		results := []fileMatches[SyntacticMatch]{}
		for _, file := range *files {
			if document, ok := documents[file.path]; ok {
				syntacticMatches, _ := findSyntacticMatchesForCandidateFile(ctx, trace, document, file)
				results = append(results, fileMatches[SyntacticMatch]{
					path:    file.path,
					matches: syntacticMatches,
				})
			}
		}
		return results, nil
	})
	if err != nil {
		return SyntacticUsagesResult{}, PreviousSyntacticSearch{}, &SyntacticUsagesError{
			Code:            SU_Fatal,
			UnderlyingError: err,
		}
	}
	finalMatches, searchedFiles := applyLimit(args.Limit, results)
	if !searchResult.limitHit && len(searchedFiles) == len(results) {
		// We're done!
		// cursor = mkSearchBasedCursor()
	} else {
		_ = searchedFiles
		// cursor = appendCursor(args.Cursor, searchedFiles)
	}
	return SyntacticUsagesResult{Matches: finalMatches}, PreviousSyntacticSearch{
		MappedIndex: mappedIndex,
		SymbolName:  symbolName,
		Language:    language,
	}, nil
}

// searchBasedUsagesImpl is extracted from SearchBasedUsages to allow
// testing of the core logic, by only mocking the search client.
func searchBasedUsagesImpl(
	ctx context.Context,
	trace observation.TraceLogger,
	searchClient searchclient.SearchClient,
	args UsagesForSymbolArgs,
	symbolName string,
	language string,
	syntacticIndex core.Option[MappedIndex],
) (_ SearchBasedUsagesResult, err error) {
	var matchResults struct {
		searchResult candidateOccurrenceResult
		err          error
	}
	var symbolResults struct {
		candidateSymbols symbolSearchResult
		err              error
	}
	mkSearchArgs := func(countLimit int32) searchArgs {
		return searchArgs{
			repo:       args.Repo.Name,
			commit:     args.Commit,
			identifier: symbolName,
			language:   language,
			countLimit: countLimit,
		}
	}
	var wg conc.WaitGroup
	wg.Go(func() {
		// TODO: Assumes at least every fifth match is a search-based one (might not hold up?)
		searchLimit := args.Limit * 5
		// If we don't have a syntactic index all matches are search-based
		// usages, so we can just fetch the exact amount we need.
		if syntacticIndex.IsNone() {
			searchLimit = args.Limit
		}
		searchCoords := mkSearchArgs(searchLimit)
		matchResults.searchResult, matchResults.err = findCandidateOccurrencesViaSearch(ctx, trace, searchClient, searchCoords)
	})
	wg.Go(func() {
		// NOTE: Same hard-coded 50 the web app used to use
		searchCoords := mkSearchArgs(50)
		symbolResults.candidateSymbols, symbolResults.err = symbolSearch(ctx, trace, searchClient, searchCoords)
	})
	wg.Wait()
	if matchResults.err != nil {
		return SearchBasedUsagesResult{}, matchResults.err
	}
	if symbolResults.err != nil {
		trace.Warn("Failed to run symbol search, will not mark any search-based usages as definitions", log.Error(symbolResults.err))
	}
	searchResult := matchResults.searchResult
	candidateSymbols := symbolResults.candidateSymbols

	tasks, _ := genslices.ChunkEvery(searchResult.candidateFiles, SYNTACTIC_USAGES_DOCUMENTS_CHUNK_SIZE)
	results := conciter.Map(tasks, func(files *[]candidateFile) []fileMatches[SearchBasedMatch] {
		documents := map[core.RepoRelPath]MappedDocument{}
		if mappedIndex, ok := syntacticIndex.Get(); ok {
			paths := genslices.Map(*files, func(f candidateFile) core.RepoRelPath {
				return f.path
			})
			documentsMap, err := mappedIndex.GetDocuments(ctx, paths)
			if err == nil {
				documents = documentsMap
			}
		}
		results := []fileMatches[SearchBasedMatch]{}
		for _, file := range *files {
			var searchBasedMatches []SearchBasedMatch
			if document, ok := documents[file.path]; ok {
				_, searchBasedMatches = findSyntacticMatchesForCandidateFile(ctx, trace, document, file)
			} else {
				for _, match := range file.matches {
					searchBasedMatches = append(searchBasedMatches, SearchBasedMatch{
						Path:               file.path,
						Range:              match.range_,
						SurroundingContent: match.surroundingContent,
						IsDefinition:       candidateSymbols.Contains(file.path, match.range_),
					})
				}
			}
			results = append(results, fileMatches[SearchBasedMatch]{
				path:    file.path,
				matches: searchBasedMatches,
			})
		}
		return results
	})

	finalMatches, searchedFiles := applyLimit(args.Limit, results)
	if !searchResult.limitHit && len(searchedFiles) == len(results) {
		// We're done!
		// cursor = None
	} else {
		_ = searchedFiles
		// cursor = appendCursor(args.Cursor, searchedFiles)
	}
	_ = searchedFiles
	return SearchBasedUsagesResult{
		Matches:    finalMatches,
		NextCursor: core.None[UsagesCursor](),
	}, nil
}

func applyLimit[T any](limit int32, fileMatchess [][]fileMatches[T]) ([]T, []core.RepoRelPath) {
	matches := make([]T, 0)
	paths := make([]core.RepoRelPath, 0)
outer:
	for _, fileMatches := range fileMatchess {
		for _, fileMatch := range fileMatches {
			if int32(len(matches)) >= limit {
				break outer
			}
			paths = append(paths, fileMatch.path)
			matches = append(matches, fileMatch.matches...)
		}
	}
	return matches, paths
}
