package graphqlbackend

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/internal/search"
	querytypes "github.com/sourcegraph/sourcegraph/internal/search/query/types"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

const maxUnindexedRepoRevSearchesPerQuery = 200

// A global limiter on number of concurrent searcher searches.
var textSearchLimiter = mutablelimiter.New(32)

// A light wrapper around the search service. We implement the service here so
// that we can unmarshal the result directly into graphql resolvers.

type FileMatch struct {
	JPath        string       `json:"Path"`
	JLineMatches []*lineMatch `json:"LineMatches"`
	JLimitHit    bool         `json:"LimitHit"`
	MatchCount   int          // Number of matches. Different from len(JLineMatches), as multiple lines may correspond to one logical match.
	symbols      []*searchSymbolResult
	uri          string
	Repo         *types.RepoName
	CommitID     api.CommitID
	// InputRev is the Git revspec that the user originally requested to search. It is used to
	// preserve the original revision specifier from the user instead of navigating them to the
	// absolute commit ID when they select a result.
	InputRev *string
}

// FileMatchResolver is a resolver for the GraphQL type `FileMatch`
type FileMatchResolver struct {
	FileMatch

	RepoResolver *RepositoryResolver
}

func (fm *FileMatchResolver) Equal(other *FileMatchResolver) bool {
	return reflect.DeepEqual(fm, other)
}

func (fm *FileMatchResolver) Key() string {
	return fm.uri
}

func (fm *FileMatchResolver) File() *GitTreeEntryResolver {
	// NOTE(sqs): Omits other commit fields to avoid needing to fetch them
	// (which would make it slow). This GitCommitResolver will return empty
	// values for all other fields.
	return &GitTreeEntryResolver{
		commit: &GitCommitResolver{
			repoResolver: fm.RepoResolver,
			oid:          GitObjectID(fm.CommitID),
			inputRev:     fm.InputRev,
		},
		stat: CreateFileInfo(fm.JPath, false),
	}
}

func (fm *FileMatchResolver) Repository() *RepositoryResolver {
	return fm.RepoResolver
}

func (fm *FileMatchResolver) RevSpec() *gitRevSpec {
	if fm.InputRev == nil || *fm.InputRev == "" {
		return nil // default branch
	}
	return &gitRevSpec{
		expr: &gitRevSpecExpr{expr: *fm.InputRev, repo: fm.Repository()},
	}
}

func (fm *FileMatchResolver) Resource() string {
	return fm.uri
}

func (fm *FileMatchResolver) Symbols() []*symbolResolver {
	symbols := make([]*symbolResolver, len(fm.symbols))
	for i, s := range fm.symbols {
		symbols[i] = toSymbolResolver(s.symbol, s.baseURI, s.lang, s.commit)
	}
	return symbols
}

func (fm *FileMatchResolver) LineMatches() []*lineMatch {
	return fm.JLineMatches
}

func (fm *FileMatchResolver) LimitHit() bool {
	return fm.JLimitHit
}

func (fm *FileMatchResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (fm *FileMatchResolver) ToFileMatch() (*FileMatchResolver, bool)   { return fm, true }
func (fm *FileMatchResolver) ToCommitSearchResult() (*CommitSearchResultResolver, bool) {
	return nil, false
}

// path returns the path in repository for the file. This isn't directly
// exposed in the GraphQL API (we expose a URI), but is used a lot internally.
func (fm *FileMatchResolver) path() string {
	return fm.JPath
}

// appendMatches appends the line matches from src as well as updating match
// counts and limit.
func (fm *FileMatchResolver) appendMatches(src *FileMatchResolver) {
	fm.JLineMatches = append(fm.JLineMatches, src.JLineMatches...)
	fm.symbols = append(fm.symbols, src.symbols...)
	fm.MatchCount += src.MatchCount
	fm.JLimitHit = fm.JLimitHit || src.JLimitHit
}

func (fm *FileMatchResolver) ResultCount() int32 {
	rc := len(fm.symbols) + fm.MatchCount
	if rc > 0 {
		return int32(rc)
	}
	return 1 // 1 to count "empty" results like type:path results
}

// lineMatch is the struct used by vscode to receive search results for a line
type lineMatch struct {
	JPreview          string     `json:"Preview"`
	JOffsetAndLengths [][2]int32 `json:"OffsetAndLengths"`
	JLineNumber       int32      `json:"LineNumber"`
	JLimitHit         bool       `json:"LimitHit"`
}

func (lm *lineMatch) Preview() string {
	return lm.JPreview
}

func (lm *lineMatch) LineNumber() int32 {
	return lm.JLineNumber
}

func (lm *lineMatch) OffsetAndLengths() [][]int32 {
	r := make([][]int32, len(lm.JOffsetAndLengths))
	for i := range lm.JOffsetAndLengths {
		r[i] = lm.JOffsetAndLengths[i][:]
	}
	return r
}

func (lm *lineMatch) LimitHit() bool {
	return lm.JLimitHit
}

var mockSearchFilesInRepo func(ctx context.Context, repo *types.RepoName, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []*FileMatchResolver, limitHit bool, err error)

func searchFilesInRepo(ctx context.Context, searcherURLs *endpoint.Map, repo *types.RepoName, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) ([]*FileMatchResolver, bool, error) {
	if mockSearchFilesInRepo != nil {
		return mockSearchFilesInRepo(ctx, repo, gitserverRepo, rev, info, fetchTimeout)
	}

	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commit, err := git.ResolveRevision(ctx, gitserverRepo, rev, git.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return nil, false, err
	}

	shouldBeSearched, err := repoShouldBeSearched(ctx, searcherURLs, info, gitserverRepo, commit, fetchTimeout)
	if err != nil {
		return nil, false, err
	}
	if !shouldBeSearched {
		return nil, false, err
	}

	var indexerEndpoints []string
	if info.IsStructuralPat {
		endpoints, err := search.Indexers().Map.Endpoints()
		for key := range endpoints {
			indexerEndpoints = append(indexerEndpoints, key)
		}
		if err != nil {
			return nil, false, err
		}
	}
	matches, limitHit, err := searcher.Search(ctx, searcherURLs, gitserverRepo, commit, info, fetchTimeout, indexerEndpoints)
	if err != nil {
		return nil, false, err
	}

	workspace := fileMatchURI(repo.Name, rev, "")
	repoResolver := &RepositoryResolver{innerRepo: repo.ToRepo()}
	resolvers := make([]*FileMatchResolver, 0, len(matches))
	for _, fm := range matches {
		lineMatches := make([]*lineMatch, 0, len(fm.LineMatches))
		for _, lm := range fm.LineMatches {
			ranges := make([][2]int32, 0, len(lm.OffsetAndLengths))
			for _, ol := range lm.OffsetAndLengths {
				ranges = append(ranges, [2]int32{int32(ol[0]), int32(ol[1])})
			}
			lineMatches = append(lineMatches, &lineMatch{
				JPreview:          lm.Preview,
				JOffsetAndLengths: ranges,
				JLineNumber:       int32(lm.LineNumber),
				JLimitHit:         lm.LimitHit,
			})
		}

		resolvers = append(resolvers, &FileMatchResolver{
			FileMatch: FileMatch{
				JPath:        fm.Path,
				JLineMatches: lineMatches,
				JLimitHit:    fm.LimitHit,
				MatchCount:   fm.MatchCount,
				Repo:         repo,
				uri:          workspace + fm.Path,
				CommitID:     commit,
				InputRev:     &rev,
			},
			RepoResolver: repoResolver,
		})
	}

	return resolvers, limitHit, err
}

// repoShouldBeSearched determines whether a repository should be searched in, based on whether the repository
// fits in the subset of repositories specified in the query's `repohasfile` and `-repohasfile` flags if they exist.
func repoShouldBeSearched(ctx context.Context, searcherURLs *endpoint.Map, searchPattern *search.TextPatternInfo, gitserverRepo api.RepoName, commit api.CommitID, fetchTimeout time.Duration) (shouldBeSearched bool, err error) {
	shouldBeSearched = true
	flagInQuery := len(searchPattern.FilePatternsReposMustInclude) > 0
	if flagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, true, searchPattern.FilePatternsReposMustInclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	negFlagInQuery := len(searchPattern.FilePatternsReposMustExclude) > 0
	if negFlagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, false, searchPattern.FilePatternsReposMustExclude, gitserverRepo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	return shouldBeSearched, nil
}

// repoHasFilesWithNamesMatching searches in a repository for matches for the patterns in the `repohasfile` or `-repohasfile` flags, and returns
// whether or not the repoShouldBeSearched in or not, based on whether matches were returned.
func repoHasFilesWithNamesMatching(ctx context.Context, searcherURLs *endpoint.Map, include bool, repoHasFileFlag []string, gitserverRepo api.RepoName, commit api.CommitID, fetchTimeout time.Duration) (bool, error) {
	for _, pattern := range repoHasFileFlag {
		p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: 1, IncludePatterns: []string{pattern}, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		matches, _, err := searcher.Search(ctx, searcherURLs, gitserverRepo, commit, &p, fetchTimeout, []string{})
		if err != nil {
			return false, err
		}
		if include && len(matches) == 0 || !include && len(matches) > 0 {
			// repo shouldn't be searched if it does not have matches for the patterns in `repohasfile`
			// or if it has file matches for the patterns in `-repohasfile`.
			return false, nil
		}
	}

	return true, nil
}

func fileMatchURI(name api.RepoName, ref, path string) string {
	var b strings.Builder
	ref = url.QueryEscape(ref)
	b.Grow(len(name) + len(ref) + len(path) + len("git://?#"))
	b.WriteString("git://")
	b.WriteString(string(name))
	if ref != "" {
		b.WriteByte('?')
		b.WriteString(ref)
	}
	b.WriteByte('#')
	b.WriteString(path)
	return b.String()
}

var mockSearchFilesInRepos func(args *search.TextParameters) ([]*FileMatchResolver, *streaming.Stats, error)

func fileMatchResultsToSearchResults(results []*FileMatchResolver) []SearchResultResolver {
	results2 := make([]SearchResultResolver, len(results))
	for i, result := range results {
		results2[i] = result
	}
	return results2
}

func searchResultsToFileMatchResults(results []SearchResultResolver) ([]*FileMatchResolver, error) {
	results2 := make([]*FileMatchResolver, len(results))
	for i, result := range results {
		fm, ok := result.ToFileMatch()
		if !ok {
			return nil, fmt.Errorf("expected only file match results")
		}
		results2[i] = fm
	}
	return results2, nil
}

// searchFilesInRepoBatch is a convenience function around searchFilesInRepos
// which collects the results from the stream.
func searchFilesInReposBatch(ctx context.Context, args *search.TextParameters) ([]*FileMatchResolver, streaming.Stats, error) {
	ctx, stream, done := collectStream(ctx)
	searchFilesInRepos(ctx, args, stream)
	agg := done()
	results, err := searchResultsToFileMatchResults(agg.Results)
	if err != nil {
		agg.Error = errors.Wrap(err, "searchFilesInReposBatch failed to convert results")
	}
	return results, agg.Stats, agg.Error
}

// searchFilesInRepos searches a set of repos for a pattern.
func searchFilesInRepos(ctx context.Context, args *search.TextParameters, stream SearchStream) {
	var (
		wg sync.WaitGroup

		mu                sync.Mutex
		searchErr         error
		resultCount       int
		overLimitCanceled bool // canceled because we were over the limit
	)
	if mockSearchFilesInRepos != nil {
		results, mockStats, err := mockSearchFilesInRepos(args)
		stream <- SearchEvent{
			Results: fileMatchResultsToSearchResults(results),
			Stats:   statsDeref(mockStats),
			Error:   err,
		}
		return
	}

	tr, ctx := trace.New(ctx, "searchFilesInRepos", fmt.Sprintf("query: %s", args.PatternInfo.Pattern))
	defer func() {
		mu.Lock()
		if searchErr != nil {
			stream <- SearchEvent{
				Error: searchErr,
			}
		}
		tr.SetError(searchErr)
		tr.Finish()
		mu.Unlock()
	}()
	fields := querytypes.Fields(args.Query.Fields())
	tr.LogFields(
		trace.Stringer("query", &fields),
		trace.Stringer("info", args.PatternInfo),
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	indexedTyp := textRequest
	if args.PatternInfo.IsStructuralPat {
		// Structural Patterns queries zoekt just file files to reduce the set
		// of files it searches.
		indexedTyp = fileRequest
	}

	// performance: for global searches, we avoid calling newIndexedSearchRequest
	// because zoekt will anyway have to search all its shards.
	var indexed *indexedSearchRequest
	if args.Mode == search.ZoektGlobalSearch {
		indexed = &indexedSearchRequest{
			args:  args,
			typ:   indexedTyp,
			repos: &indexedRepoRevs{},
		}
	} else {
		var err error
		indexed, err = newIndexedSearchRequest(ctx, args, indexedTyp)
		if err != nil {
			mu.Lock()
			searchErr = err
			mu.Unlock()
			return
		}
	}

	// if there are no indexed repos and this is a structural search
	// query, there will be no results. Raise a friendly alert.
	if args.PatternInfo.IsStructuralPat && len(indexed.Repos()) == 0 {
		mu.Lock()
		searchErr = errors.New("no indexed repositories for structural search")
		mu.Unlock()
		return
	}

	if args.PatternInfo.IsEmpty() {
		// Empty query isn't an error, but it has no results.
		return
	}

	tr.LazyPrintf("%d indexed repos, %d unindexed repos", len(indexed.Repos()), len(indexed.Unindexed))

	var searcherRepos []*search.RepositoryRevisions
	if indexed.DisableUnindexedSearch {
		tr.LazyPrintf("disabling unindexed search")
		var status search.RepoStatusMap
		for _, r := range indexed.Unindexed {
			status.Update(r.Repo.ID, search.RepoStatusMissing)
		}
		stream <- SearchEvent{
			Stats: streaming.Stats{
				Status: status,
			},
		}
	} else {
		// Limit the number of unindexed repositories searched for a single
		// query. Searching more than this will merely flood the system and
		// network with requests that will timeout.
		var missing []*types.RepoName
		searcherRepos, missing = limitSearcherRepos(indexed.Unindexed, maxUnindexedRepoRevSearchesPerQuery)
		if len(missing) > 0 {
			tr.LazyPrintf("limiting unindexed repos searched to %d", maxUnindexedRepoRevSearchesPerQuery)
			var status search.RepoStatusMap
			for _, r := range missing {
				status.Update(r.ID, search.RepoStatusMissing)
			}
			stream <- SearchEvent{
				Stats: streaming.Stats{
					Status: status,
				},
			}
		}
	}

	// send assumes the caller does not hold mu.
	send := func(ctx context.Context, source fmt.Stringer, event SearchEvent) {
		// Do not pass on errors yet.
		if event.Error != nil {
			if ctx.Err() == context.Canceled {
				// Our request has been canceled (another backend had a fatal
				// error, or otherwise), so we can just ignore these
				// results.
				return
			}

			// Check if we are the first error found.
			mu.Lock()
			if searchErr == nil && !overLimitCanceled {
				searchErr = errors.Wrapf(event.Error, "failed to search %s", source.String())
				tr.LazyPrintf("cancel due to error: %v", searchErr)
				cancel()
			}
			mu.Unlock()

			// Do not report the error now on the stream. We report a final
			// error (searchErr) once all backends have finished running.
			return
		}

		stream <- event

		// Stop searching if we have found enough results.
		mu.Lock()
		resultCount += len(event.Results)
		if limit := int(args.PatternInfo.FileMatchLimit); resultCount > limit && !overLimitCanceled {
			cancel()
			tr.LazyPrintf("cancel due to result size: %d > %d", resultCount, limit)
			overLimitCanceled = true

			// Inform stream we have found more than limit results
			stream <- SearchEvent{
				Stats: streaming.Stats{
					IsLimitHit: true,
				},
			}
		}
		mu.Unlock()
	}

	// callSearcherOverRepos calls searcher on a set of repos.
	// searcherReposFilteredFiles is an optional map of {repo name => file list}
	// that forces the searcher to only include the file list in the
	// search. It is currently only set when Zoekt restricts the file list for structural search.
	callSearcherOverRepos := func(
		searcherRepos []*search.RepositoryRevisions,
		searcherReposFilteredFiles map[string][]string,
	) error {
		var fetchTimeout time.Duration
		if len(searcherRepos) == 1 || args.UseFullDeadline {
			// When searching a single repo or when an explicit timeout was specified, give it the remaining deadline to fetch the archive.
			deadline, ok := ctx.Deadline()
			if ok {
				fetchTimeout = time.Until(deadline)
			} else {
				// In practice, this case should not happen because a deadline should always be set
				// but if it does happen just set a long but finite timeout.
				fetchTimeout = time.Minute
			}
		} else {
			// When searching many repos, don't wait long for any single repo to fetch.
			fetchTimeout = 500 * time.Millisecond
		}

		if len(searcherRepos) > 0 {
			// The number of searcher endpoints can change over time. Inform our
			// limiter of the new limit, which is a multiple of the number of
			// searchers.
			eps, err := args.SearcherURLs.Endpoints()
			if err != nil {
				return err
			}
			textSearchLimiter.SetLimit(len(eps) * 32)
		}

	outer:
		for _, repoAllRevs := range searcherRepos {
			if len(repoAllRevs.Revs) == 0 {
				continue
			}

			revSpecs, err := repoAllRevs.ExpandedRevSpecs(ctx)
			if err != nil {
				return err
			}

			for _, rev := range revSpecs {
				// Only reason acquire can fail is if ctx is cancelled. So we can stop
				// looping through searcherRepos.
				limitCtx, limitDone, acquireErr := textSearchLimiter.Acquire(ctx)
				if acquireErr != nil {
					break outer
				}

				// Make a new repoRev for just the operation of searching this revspec.
				repoRev := &search.RepositoryRevisions{Repo: repoAllRevs.Repo, Revs: []search.RevisionSpecifier{{RevSpec: rev}}}

				args := *args
				if args.PatternInfo.IsStructuralPat && searcherReposFilteredFiles != nil {
					// Modify the search query to only run for the filtered files
					if v, ok := searcherReposFilteredFiles[string(repoRev.Repo.Name)]; ok {
						patternCopy := *args.PatternInfo
						args.PatternInfo = &patternCopy
						includePatternsCopy := []string{}
						args.PatternInfo.IncludePatterns = append(includePatternsCopy, v...)
					}
				}

				wg.Add(1)
				go func(ctx context.Context, done context.CancelFunc) {
					defer wg.Done()
					defer done()

					matches, repoLimitHit, err := searchFilesInRepo(ctx, args.SearcherURLs, repoRev.Repo, repoRev.GitserverRepo(), repoRev.RevSpecs()[0], args.PatternInfo, fetchTimeout)
					if err != nil {
						tr.LogFields(otlog.String("repo", string(repoRev.Repo.Name)), otlog.Error(err), otlog.Bool("timeout", errcode.IsTimeout(err)), otlog.Bool("temporary", errcode.IsTemporary(err)))
						log15.Warn("searchFilesInRepo failed", "error", err, "repo", repoRev.Repo.Name)
					}
					// non-diff search reports timeout through err, so pass false for timedOut
					repoCommon, fatalErr := handleRepoSearchResult(repoRev, repoLimitHit, false, err)
					send(ctx, repoRev, SearchEvent{
						Results: fileMatchResultsToSearchResults(matches),
						Stats:   repoCommon,
						Error:   fatalErr,
					})
				}(limitCtx, limitDone) // ends the Go routine for a call to searcher for a repo
			} // ends the for loop iterating over repo's revs
		} // ends the for loop iterating over repos
		return nil
	} // ends callSearcherOverRepos

	// Indexed regex and literal search go via zoekt
	isIndexedTextSearch := args.Mode != search.SearcherOnly && !args.PatternInfo.IsStructuralPat
	// Structural search goes via zoekt then searcher.
	isStructuralSearch := args.Mode != search.SearcherOnly && args.PatternInfo.IsStructuralPat

	if isIndexedTextSearch {
		wg.Add(1)
		go func() {
			// TODO limitHit, handleRepoSearchResult
			defer wg.Done()
			for event := range indexed.Search(ctx) {
				tr.LogFields(otlog.Int("matches.len", len(event.Results)), otlog.Error(event.Error))
				send(ctx, stringerFunc("indexed"), event)
			}
		}()
	}

	if isStructuralSearch && args.PatternInfo.CombyRule == `where "zoekt" == "zoekt"` {
		wg.Add(1)
		go func() {
			defer wg.Done()

			repos := make([]*search.RepositoryRevisions, 0, len(indexed.Repos()))
			for _, repo := range indexed.Repos() {
				repos = append(repos, repo)
			}

			err := callSearcherOverRepos(repos, nil)
			if err != nil {
				mu.Lock()
				searchErr = err
				mu.Unlock()
			}
		}()
	} else if isStructuralSearch {
		wg.Add(1)
		go func() {
			// TODO limitHit, handleRepoSearchResult
			defer wg.Done()
			for event := range indexed.Search(ctx) {
				tr.LogFields(otlog.Int("matches.len", len(event.Results)), otlog.Error(event.Error))
				send(ctx, stringerFunc("structural-indexed"), SearchEvent{
					Stats: event.Stats,
				})

				// For structural search, we run callSearcherOverRepos
				// over the set of repos and files known to contain
				// parts of the pattern as determined by Zoekt.

				// A partition of {repo name => file list} that we will build from Zoekt matches
				partition := make(map[string][]string)
				var repos []*search.RepositoryRevisions

				for _, m := range event.Results {
					fm, ok := m.ToFileMatch()
					if !ok {
						mu.Lock()
						searchErr = fmt.Errorf("structual search: Events from indexed.Search could not be converted to FileMatch")
						mu.Unlock()
						return
					}
					name := string(fm.Repo.Name)
					partition[name] = append(partition[name], fm.JPath)
				}

				// Filter Zoekt repos that didn't contain matches
				for _, repo := range indexed.Repos() {
					for key := range partition {
						if string(repo.Repo.Name) == key {
							repos = append(repos, repo)
						}
					}
				}

				err := callSearcherOverRepos(repos, partition)
				if err != nil {
					mu.Lock()
					searchErr = err
					mu.Unlock()
				}
			}
		}()
	}

	// This guard disables
	// - unindexed structural search
	// - unindexed search of negated content
	if !args.PatternInfo.IsStructuralPat {
		if err := callSearcherOverRepos(searcherRepos, nil); err != nil {
			mu.Lock()
			searchErr = err
			mu.Unlock()
		}
	}

	wg.Wait()
}

// limitSearcherRepos limits the number of repo@revs searched by the unindexed searcher codepath.
// Sending many requests to searcher would otherwise cause a flood of system and network requests
// that result in timeouts or long delays.
//
// It returns the new repositories destined for the unindexed searcher code path, and the
// repositories that are limited / excluded.
//
// A slice to the input list is returned, it is not copied.
func limitSearcherRepos(unindexed []*search.RepositoryRevisions, limit int) (searcherRepos []*search.RepositoryRevisions, limitedSearcherRepos []*types.RepoName) {
	totalRepoRevs := 0
	limitedRepos := 0
	for _, repoRevs := range unindexed {
		totalRepoRevs += len(repoRevs.Revs)
		if totalRepoRevs > limit {
			limitedSearcherRepos = append(limitedSearcherRepos, repoRevs.Repo)
			limitedRepos++
		}
	}
	searcherRepos = unindexed[:len(unindexed)-limitedRepos]
	return
}

type stringerFunc string

func (s stringerFunc) String() string {
	return string(s)
}
