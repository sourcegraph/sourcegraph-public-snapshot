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

type TextSearchJob struct {
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

	Features search.Features
}

// Run calls the searcher service on a set of repositories.
func (s *TextSearchJob) Run(ctx context.Context, clients job.RuntimeClients, stream streaming.Sender) (alert *search.Alert, err error) {
	tr, ctx, stream, finish := job.StartSpan(ctx, stream, s)
	defer func() { finish(alert, err) }()

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

					repoLimitHit, err := s.searchFilesInRepo(ctx, clients.DB, clients.SearcherURLs, repoRev.Repo, repoRev.GitserverRepo(), repoRev.RevSpecs()[0], s.Indexed, s.PatternInfo, fetchTimeout, stream)
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

func (s *TextSearchJob) Name() string {
	return "SearcherTextSearchJob"
}

func (s *TextSearchJob) Tags() []log.Field {
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

func (s *TextSearchJob) searchFilesInRepo(
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

	// Structural and hybrid search both speak to zoekt so need the endpoints.
	var indexerEndpoints []string
	if info.IsStructuralPat || s.Features.HybridSearch {
		indexerEndpoints, err = search.Indexers().Map.Endpoints()
		if err != nil {
			return false, err
		}
	}

	onMatches := func(searcherMatches []*protocol.FileMatch) {
		stream.Send(streaming.SearchEvent{
			Results: convertMatches(repo, commit, &rev, searcherMatches),
		})
	}

	return Search(ctx, searcherURLs, gitserverRepo, repo.ID, rev, commit, index, info, fetchTimeout, indexerEndpoints, s.Features, onMatches)
}

// convert converts a set of searcher matches into []result.Match
func convertMatches(repo types.MinimalRepo, commit api.CommitID, rev *string, searcherMatches []*protocol.FileMatch) []result.Match {
	matches := make([]result.Match, 0, len(searcherMatches))
	for _, fm := range searcherMatches {
		chunkMatches := make(result.ChunkMatches, 0, len(fm.ChunkMatches))

		for _, cm := range fm.ChunkMatches {
			ranges := make(result.Ranges, 0, len(cm.Ranges))
			for _, rr := range cm.Ranges {
				ranges = append(ranges, result.Range{
					Start: result.Location{
						Offset: int(rr.Start.Offset),
						Line:   int(rr.Start.Line),
						Column: int(rr.Start.Column),
					},
					End: result.Location{
						Offset: int(rr.End.Offset),
						Line:   int(rr.End.Line),
						Column: int(rr.End.Column),
					},
				})
			}

			chunkMatches = append(chunkMatches, result.ChunkMatch{
				Content: cm.Content,
				ContentStart: result.Location{
					Offset: int(cm.ContentStart.Offset),
					Line:   int(cm.ContentStart.Line),
					Column: 0,
				},
				Ranges: ranges,
			})
		}

		matches = append(matches, &result.FileMatch{
			File: result.File{
				Path:     fm.Path,
				Repo:     repo,
				CommitID: commit,
				InputRev: rev,
			},
			ChunkMatches: chunkMatches,
			LimitHit:     fm.LimitHit,
		})
	}
	return matches
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
		// TODO(keegancsmith) we should be passing in more state here like
		// indexer endpoints and features.
		p := search.TextPatternInfo{IsRegExp: true, FileMatchLimit: 1, IncludePatterns: []string{pattern}, PathPatternsAreCaseSensitive: false, PatternMatchesContent: true, PatternMatchesPath: true}
		_, err := Search(ctx, searcherURLs, repo.Name, repo.ID, "", commit, false, &p, fetchTimeout, []string{}, search.Features{}, onMatches)
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
