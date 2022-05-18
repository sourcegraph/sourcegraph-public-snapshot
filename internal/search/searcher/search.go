package searcher

import (
	"context"
	"time"

	"github.com/inconshreveable/log15"
	"github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/cmd/searcher/protocol"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/endpoint"
	"github.com/sourcegraph/sourcegraph/internal/errcode"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/mutablelimiter"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/job"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// A global limiter on number of concurrent searcher searches.
var textSearchLimiter = mutablelimiter.New(32)

type SearcherJob struct {
	PatternInfo *search.TextPatternInfo
	Repos       []*search.RepositoryRevisions // the set of repositories to search with searcher.

	// Indexed represents whether the set of repositories are indexed (used
	// to communicate whether searcher should call Zoekt search on these
	// repos).
	Indexed bool

	// UseFullDeadline indicates that the search should try do as much work as
	// it can within context.Deadline. If false the search should try and be
	// as fast as possible, even if a "slow" deadline is set.
	//
	// For example searcher will wait to full its archive cache for a
	// repository if this field is true. Another example is we set this field
	// to true if the user requests a specific timeout or maximum result size.
	UseFullDeadline bool
}

// Run calls the searcher service on a set of repositories.
func (s *SearcherJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()
	tr.TagFields(trace.LazyFields(s.Tags))

	var fetchTimeout time.Duration
	if len(s.Repos) == 1 || s.UseFullDeadline {
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
		log.Int64("fetch_timeout_ms", fetchTimeout.Milliseconds()),
		log.Int64("repos_count", int64(len(s.Repos))),
	)

	if len(s.Repos) == 0 {
		return nil, nil
	}

	// The number of searcher endpoints can change over time. Inform our
	// limiter of the new limit, which is a multiple of the number of
	// searchers.
	eps, err := clients.SearcherURLs.Endpoints()
	if err != nil {
		return nil, err
	}
	textSearchLimiter.SetLimit(len(eps) * 32)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		for _, repoAllRevs := range s.Repos {
			if len(repoAllRevs.Revs) == 0 {
				continue
			}

			revSpecs, err := repoAllRevs.ExpandedRevSpecs(ctx, clients.DB)
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

					repoLimitHit, err := searchFilesInRepo(ctx, clients.DB, clients.SearcherURLs, repoRev.Repo, repoRev.GitserverRepo(), repoRev.RevSpecs()[0], s.Indexed, s.PatternInfo, fetchTimeout, stream)
					if err != nil {
						tr.LogFields(log.String("repo", string(repoRev.Repo.Name)), log.Error(err), log.Bool("timeout", errcode.IsTimeout(err)), log.Bool("temporary", errcode.IsTemporary(err)))
						log15.Warn("searchFilesInRepo failed", "error", err, "repo", repoRev.Repo.Name)
					}
					// non-diff search reports timeout through err, so pass false for timedOut
					status, limitHit, err := search.HandleRepoSearchResult(repoRev, repoLimitHit, false, err)
					stream.Send(streaming.SearchEvent{
						Stats: streaming.Stats{
							Status:     status,
							IsLimitHit: limitHit,
						},
					})
					return err
				})
			}
		}

		return nil
	})

	return nil, g.Wait()
}

func (s *SearcherJob) Name() string {
	return "SearcherJob"
}

func (s *SearcherJob) Tags() []log.Field {
	return []log.Field{
		trace.Stringer("patternInfo", s.PatternInfo),
		log.Int("numRepos", len(s.Repos)),
		log.Bool("indexed", s.Indexed),
		log.Bool("useFullDeadline", s.UseFullDeadline),
	}
}

var MockSearchFilesInRepo func(
	ctx context.Context,
	repo types.MinimalRepo,
	gitserverRepo api.RepoName,
	rev string,
	info *search.TextPatternInfo,
	fetchTimeout time.Duration,
	stream streaming.Sender,
) (limitHit bool, err error)

func searchFilesInRepo(
	ctx context.Context,
	db database.DB,
	searcherURLs *endpoint.Map,
	repo types.MinimalRepo,
	gitserverRepo api.RepoName,
	rev string,
	index bool,
	info *search.TextPatternInfo,
	fetchTimeout time.Duration,
	stream streaming.Sender,
) (bool, error) {
	if MockSearchFilesInRepo != nil {
		return MockSearchFilesInRepo(ctx, repo, gitserverRepo, rev, info, fetchTimeout, stream)
	}

	// Do not trigger a repo-updater lookup (e.g.,
	// backend.{GitRepo,Repos.ResolveRev}) because that would slow this operation
	// down by a lot (if we're looping over many repos). This means that it'll fail if a
	// repo is not on gitserver.
	commit, err := gitserver.NewClient(db).ResolveRevision(ctx, gitserverRepo, rev, gitserver.ResolveRevisionOptions{NoEnsureRevision: true})
	if err != nil {
		return false, err
	}

	shouldBeSearched, err := repoShouldBeSearched(ctx, searcherURLs, info, repo, commit, fetchTimeout)
	if err != nil {
		return false, err
	}
	if !shouldBeSearched {
		return false, err
	}

	var indexerEndpoints []string
	if info.IsStructuralPat {
		indexerEndpoints, err = search.Indexers().Map.Endpoints()
		if err != nil {
			return false, err
		}
	}

	toMatches := newToMatches(repo, commit, &rev)
	onMatches := func(searcherMatches []*protocol.FileMatch) {
		stream.Send(streaming.SearchEvent{
			Results: toMatches(searcherMatches),
		})
	}

	return Search(ctx, searcherURLs, gitserverRepo, repo.ID, rev, commit, index, info, fetchTimeout, indexerEndpoints, onMatches)
}

// newToMatches returns a closure that converts []*protocol.FileMatch to []result.Match.
func newToMatches(repo types.MinimalRepo, commit api.CommitID, rev *string) func([]*protocol.FileMatch) []result.Match {
	return func(searcherMatches []*protocol.FileMatch) []result.Match {
		matches := make([]result.Match, 0, len(searcherMatches))
		for _, fm := range searcherMatches {
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

			matches = append(matches, &result.FileMatch{
				File: result.File{
					Path:     fm.Path,
					Repo:     repo,
					CommitID: commit,
					InputRev: rev,
				},
				LineMatches: lineMatches,
				LimitHit:    fm.LimitHit,
			})
		}
		return matches
	}
}

// repoShouldBeSearched determines whether a repository should be searched in, based on whether the repository
// fits in the subset of repositories specified in the query's `repohasfile` and `-repohasfile` flags if they exist.
func repoShouldBeSearched(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	searchPattern *search.TextPatternInfo,
	repo types.MinimalRepo,
	commit api.CommitID,
	fetchTimeout time.Duration,
) (shouldBeSearched bool, err error) {
	shouldBeSearched = true
	flagInQuery := len(searchPattern.FilePatternsReposMustInclude) > 0
	if flagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, true, searchPattern.FilePatternsReposMustInclude, repo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	negFlagInQuery := len(searchPattern.FilePatternsReposMustExclude) > 0
	if negFlagInQuery {
		shouldBeSearched, err = repoHasFilesWithNamesMatching(ctx, searcherURLs, false, searchPattern.FilePatternsReposMustExclude, repo, commit, fetchTimeout)
		if err != nil {
			return shouldBeSearched, err
		}
	}
	return shouldBeSearched, nil
}

// repoHasFilesWithNamesMatching searches in a repository for matches for the patterns in the `repohasfile` or `-repohasfile` flags, and returns
// whether or not the repoShouldBeSearched in or not, based on whether matches were returned.
func repoHasFilesWithNamesMatching(
	ctx context.Context,
	searcherURLs *endpoint.Map,
	include bool,
	repoHasFileFlag []string,
	repo types.MinimalRepo,
	commit api.CommitID,
	fetchTimeout time.Duration,
) (bool, error) {
	for _, pattern := range repoHasFileFlag {
		foundMatches := false
		onMatches := func(matches []*protocol.FileMatch) {
			if len(matches) > 0 {
				foundMatches = true
			}
		}
		p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: 1, IncludePatterns: []string{pattern}, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		_, err := Search(ctx, searcherURLs, repo.Name, repo.ID, "", commit, false, &p, fetchTimeout, []string{}, onMatches)
		if err != nil {
			return false, err
		}
		if include && !foundMatches || !include && foundMatches {
			// repo shouldn't be searched if it does not have matches for the patterns in `repohasfile`
			// or if it has file matches for the patterns in `-repohasfile`.
			return false, nil
		}
	}

	return true, nil
}
