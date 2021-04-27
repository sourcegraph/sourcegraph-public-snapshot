package graphqlbackend

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"

	otlog "github.com/opentracing/opentracing-go/log"

	"github.com/inconshreveable/log15"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/filter"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/internal/vcs/git"
)

const maxUnindexedRepoRevSearchesPerQuery = 200

// A global limiter on number of concurrent searcher searches.
var textSearchLimiter = mutablelimiter.New(32)

// FileMatchResolver is a resolver for the GraphQL type `FileMatch`
type FileMatchResolver struct {
	result.FileMatch

	RepoResolver *RepositoryResolver
	db           dbutil.DB
}

// Equal provides custom comparison which is used by go-cmp
func (fm *FileMatchResolver) Equal(other *FileMatchResolver) bool {
	return reflect.DeepEqual(fm, other)
}

func (fm *FileMatchResolver) Key() string {
	return fm.URL()
}

func (fm *FileMatchResolver) File() *GitTreeEntryResolver {
	// NOTE(sqs): Omits other commit fields to avoid needing to fetch them
	// (which would make it slow). This GitCommitResolver will return empty
	// values for all other fields.
	return &GitTreeEntryResolver{
		db:     fm.db,
		commit: fm.Commit(),
		stat:   CreateFileInfo(fm.Path, false),
	}
}

func (fm *FileMatchResolver) Commit() *GitCommitResolver {
	return &GitCommitResolver{
		db:           fm.db,
		repoResolver: fm.RepoResolver,
		oid:          GitObjectID(fm.CommitID),
		inputRev:     fm.InputRev,
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
	return fm.URL()
}

func (fm *FileMatchResolver) Symbols() []symbolResolver {
	return symbolResultsToResolvers(fm.db, fm.Commit(), fm.FileMatch.Symbols)
}

func (fm *FileMatchResolver) LineMatches() []lineMatchResolver {
	r := make([]lineMatchResolver, 0, len(fm.FileMatch.LineMatches))
	for _, lm := range fm.FileMatch.LineMatches {
		r = append(r, lineMatchResolver{lm})
	}
	return r
}

func (fm *FileMatchResolver) LimitHit() bool {
	return fm.FileMatch.LimitHit
}

func (fm *FileMatchResolver) ToRepository() (*RepositoryResolver, bool) { return nil, false }
func (fm *FileMatchResolver) ToFileMatch() (*FileMatchResolver, bool)   { return fm, true }
func (fm *FileMatchResolver) ToCommitSearchResult() (*CommitSearchResultResolver, bool) {
	return nil, false
}

// path returns the path in repository for the file. This isn't directly
// exposed in the GraphQL API (we expose a URI), but is used a lot internally.
func (fm *FileMatchResolver) path() string {
	return fm.Path
}

// appendMatches appends the line matches from src as well as updating match
// counts and limit.
func (fm *FileMatchResolver) appendMatches(src *FileMatchResolver) {
	fm.FileMatch.AppendMatches(&src.FileMatch)
}

func (fm *FileMatchResolver) ResultCount() int32 {
	return int32(fm.FileMatch.ResultCount())
}

func (fm *FileMatchResolver) Select(t filter.SelectPath) SearchResultResolver {
	match := fm.FileMatch.Select(t)

	// Turn the result type back to a resolver
	switch v := match.(type) {
	case *result.RepoMatch:
		return NewRepositoryResolver(fm.db, &types.Repo{Name: v.Name, ID: v.ID})
	case *result.FileMatch:
		return &FileMatchResolver{db: fm.db, RepoResolver: fm.RepoResolver, FileMatch: *v}
	}

	return nil
}

type lineMatchResolver struct {
	*result.LineMatch
}

func (lm lineMatchResolver) Preview() string {
	return lm.LineMatch.Preview
}

func (lm lineMatchResolver) LineNumber() int32 {
	return lm.LineMatch.LineNumber
}

func (lm lineMatchResolver) OffsetAndLengths() [][]int32 {
	r := make([][]int32, len(lm.LineMatch.OffsetAndLengths))
	for i := range lm.LineMatch.OffsetAndLengths {
		r[i] = lm.LineMatch.OffsetAndLengths[i][:]
	}
	return r
}

func (lm lineMatchResolver) LimitHit() bool {
	return false
}

var mockSearchFilesInRepo func(ctx context.Context, repo types.RepoName, gitserverRepo api.RepoName, rev string, info *search.TextPatternInfo, fetchTimeout time.Duration) (matches []result.FileMatch, limitHit bool, err error)

func searchFilesInRepo(ctx context.Context, searcherURLs *endpoint.Map, repo types.RepoName, gitserverRepo api.RepoName, rev string, index bool, info *search.TextPatternInfo, fetchTimeout time.Duration) ([]result.FileMatch, bool, error) {
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
	matches, limitHit, err := searcher.Search(ctx, searcherURLs, gitserverRepo, rev, commit, index, info, fetchTimeout, indexerEndpoints)
	if err != nil {
		return nil, false, err
	}

	fileMatches := make([]result.FileMatch, 0, len(matches))
	for _, fm := range matches {
		lineMatches := make([]*result.LineMatch, 0, len(fm.LineMatches))
		for _, lm := range fm.LineMatches {
			ranges := make([][2]int32, 0, len(lm.OffsetAndLengths))
			for _, ol := range lm.OffsetAndLengths {
				ranges = append(ranges, [2]int32{int32(ol[0]), int32(ol[1])})
			}
			lineMatches = append(lineMatches, &result.LineMatch{
				Preview:          lm.Preview,
				OffsetAndLengths: ranges,
				LineNumber:       int32(lm.LineNumber),
			})
		}

		fileMatches = append(fileMatches, result.FileMatch{
			Path:        fm.Path,
			LineMatches: lineMatches,
			LimitHit:    fm.LimitHit,
			Repo:        repo,
			CommitID:    commit,
			InputRev:    &rev,
		})
	}

	return fileMatches, limitHit, err
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
		matches, _, err := searcher.Search(ctx, searcherURLs, gitserverRepo, "", commit, false, &p, fetchTimeout, []string{})
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

var mockSearchFilesInRepos func(args *search.TextParameters) ([]*FileMatchResolver, *streaming.Stats, error)

func fileMatchesToSearchResults(db dbutil.DB, matches []result.FileMatch) []SearchResultResolver {
	results := make([]SearchResultResolver, len(matches))
	for i, match := range matches {
		results[i] = &FileMatchResolver{
			FileMatch:    match,
			RepoResolver: NewRepositoryResolver(db, match.Repo.ToRepo()),
			db:           db,
		}
	}
	return results
}

func fileMatchResolversToSearchResults(resolvers []*FileMatchResolver) []SearchResultResolver {
	results := make([]SearchResultResolver, len(resolvers))
	for i, resolver := range resolvers {
		results[i] = resolver
	}
	return results
}

func searchResultsToFileMatchResults(resolvers []SearchResultResolver) ([]*FileMatchResolver, error) {
	results := make([]*FileMatchResolver, len(resolvers))
	for i, resolver := range resolvers {
		fm, ok := resolver.ToFileMatch()
		if !ok {
			return nil, fmt.Errorf("expected only file match results")
		}
		results[i] = fm
	}
	return results, nil
}

// searchFilesInRepoBatch is a convenience function around searchFilesInRepos
// which collects the results from the stream.
func searchFilesInReposBatch(ctx context.Context, db dbutil.DB, args *search.TextParameters) ([]*FileMatchResolver, streaming.Stats, error) {
	results, stats, err := collectStream(func(stream Sender) error {
		return searchFilesInRepos(ctx, db, args, stream)
	})
	fms, fmErr := searchResultsToFileMatchResults(results)
	if fmErr != nil && err == nil {
		err = errors.Wrap(fmErr, "searchFilesInReposBatch failed to convert results")
	}
	return fms, stats, err
}

// searchFilesInRepos searches a set of repos for a pattern.
func searchFilesInRepos(ctx context.Context, db dbutil.DB, args *search.TextParameters, stream Sender) (err error) {
	if mockSearchFilesInRepos != nil {
		results, mockStats, err := mockSearchFilesInRepos(args)
		stream.Send(SearchEvent{
			Results: fileMatchResolversToSearchResults(results),
			Stats:   statsDeref(mockStats),
		})
		return err
	}

	ctx, stream, cleanup := WithLimit(ctx, stream, int(args.PatternInfo.FileMatchLimit))
	defer cleanup()

	tr, ctx := trace.New(ctx, "searchFilesInRepos", fmt.Sprintf("query: %s", args.PatternInfo.Pattern))
	defer func() {
		tr.SetErrorIfNotContext(err)
		tr.Finish()
	}()
	tr.LogFields(
		trace.Stringer("query", args.Query),
		trace.Stringer("info", args.PatternInfo),
		trace.Stringer("global_search_mode", args.Mode),
	)

	// performance: for global searches, we avoid calling newIndexedSearchRequest
	// because zoekt will anyway have to search all its shards.
	var indexed *indexedSearchRequest
	if args.Mode == search.ZoektGlobalSearch {
		indexed = &indexedSearchRequest{
			db:    db,
			args:  args,
			typ:   textRequest,
			repos: &indexedRepoRevs{},
		}
	} else {
		indexed, err = newIndexedSearchRequest(ctx, db, args, textRequest, stream)
		if err != nil {
			return err
		}
	}

	if args.PatternInfo.IsEmpty() {
		// Empty query isn't an error, but it has no results.
		return nil
	}

	g, ctx := errgroup.WithContext(ctx)

	if args.Mode != search.SearcherOnly {
		// Run searches on indexed repositories.

		if !args.PatternInfo.IsStructuralPat {
			// Run literal and regexp searches.
			g.Go(func() error {
				return indexed.Search(ctx, stream)
			})
		} else {
			// Run structural search (fulfilled via searcher).
			g.Go(func() error {
				repos := make([]*search.RepositoryRevisions, 0, len(indexed.Repos()))
				for _, repo := range indexed.Repos() {
					repos = append(repos, repo)
				}
				return callSearcherOverRepos(ctx, db, args, stream, repos, true)
			})
		}
	}

	// Concurrently run searcher for all unindexed repos regardless whether text, regexp, or structural search.
	g.Go(func() error {
		return callSearcherOverRepos(ctx, db, args, stream, indexed.Unindexed, false)
	})

	return g.Wait()
}

// callSearcherOverRepos calls searcher on searcherRepos.
func callSearcherOverRepos(
	ctx context.Context,
	db dbutil.DB,
	args *search.TextParameters,
	stream Sender,
	searcherRepos []*search.RepositoryRevisions,
	index bool,
) (err error) {
	tr, ctx := trace.New(ctx, "searcherOverRepos", fmt.Sprintf("query: %s", args.PatternInfo.Pattern))
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

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

	tr.LogFields(
		otlog.Int64("fetch_timeout_ms", fetchTimeout.Milliseconds()),
		otlog.Int64("repos_count", int64(len(searcherRepos))),
	)

	if len(searcherRepos) == 0 {
		return nil
	}

	// The number of searcher endpoints can change over time. Inform our
	// limiter of the new limit, which is a multiple of the number of
	// searchers.
	eps, err := args.SearcherURLs.Endpoints()
	if err != nil {
		return err
	}
	textSearchLimiter.SetLimit(len(eps) * 32)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for _, repoAllRevs := range searcherRepos {
			if len(repoAllRevs.Revs) == 0 {
				continue
			}

			revSpecs, err := repoAllRevs.ExpandedRevSpecs(ctx)
			if err != nil {
				return err
			}

			for _, rev := range revSpecs {
				limitCtx, limitDone, err := textSearchLimiter.Acquire(ctx)
				if err != nil {
					return err
				}

				// Make a new repoRev for just the operation of searching this revspec.
				repoRev := &search.RepositoryRevisions{Repo: repoAllRevs.Repo, Revs: []search.RevisionSpecifier{{RevSpec: rev}}}
				g.Go(func() error {
					ctx, done := limitCtx, limitDone
					defer done()

					matches, repoLimitHit, err := searchFilesInRepo(ctx, args.SearcherURLs, repoRev.Repo, repoRev.GitserverRepo(), repoRev.RevSpecs()[0], index, args.PatternInfo, fetchTimeout)
					if err != nil {
						tr.LogFields(otlog.String("repo", string(repoRev.Repo.Name)), otlog.Error(err), otlog.Bool("timeout", errcode.IsTimeout(err)), otlog.Bool("temporary", errcode.IsTemporary(err)))
						log15.Warn("searchFilesInRepo failed", "error", err, "repo", repoRev.Repo.Name)
					}
					// non-diff search reports timeout through err, so pass false for timedOut
					stats, err := handleRepoSearchResult(repoRev, repoLimitHit, false, err)
					stream.Send(SearchEvent{
						Results: fileMatchesToSearchResults(db, matches),
						Stats:   stats,
					})
					return err
				})
			}
		}

		return nil
	})

	return g.Wait()
}
