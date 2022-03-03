package run

import (
	"context"
	"math"

	otlog "github.com/opentracing/opentracing-go/log"
	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type RepoSearch struct {
	Args *search.TextParameters
}

func (s *RepoSearch) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "RepoSearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	tr.LogFields(otlog.String("pattern", s.Args.PatternInfo.Pattern))

	repos := &searchrepos.Resolver{DB: db, Opts: s.Args.RepoOptions}
	err = repos.Paginate(ctx, nil, func(page *searchrepos.Resolved) error {
		tr.LogFields(otlog.Int("resolved.len", len(page.RepoRevs)))

		// Filter the repos if there is a repohasfile: or -repohasfile field.
		if len(s.Args.PatternInfo.FilePatternsReposMustExclude) > 0 || len(s.Args.PatternInfo.FilePatternsReposMustInclude) > 0 {
			// Fallback to batch for reposToAdd
			page.RepoRevs, err = reposToAdd(ctx, s.Args, page.RepoRevs)
			if err != nil {
				return err
			}
		}

		stream.Send(streaming.SearchEvent{
			Results: repoRevsToRepoMatches(ctx, page.RepoRevs),
		})

		return nil
	})

	if errors.Is(err, searchrepos.ErrNoResolvedRepos) {
		err = nil
	}

	return nil, err
}

func (*RepoSearch) Name() string {
	return "Repo"
}

func repoRevsToRepoMatches(ctx context.Context, repos []*search.RepositoryRevisions) []result.Match {
	matches := make([]result.Match, 0, len(repos))
	for _, r := range repos {
		revs, err := r.ExpandedRevSpecs(ctx)
		if err != nil { // fallback to just return revspecs
			revs = r.RevSpecs()
		}
		for _, rev := range revs {
			matches = append(matches, &result.RepoMatch{
				Name: r.Repo.Name,
				ID:   r.Repo.ID,
				Rev:  rev,
			})
		}
	}
	return matches
}

func matchesToFileMatches(matches []result.Match) ([]*result.FileMatch, error) {
	fms := make([]*result.FileMatch, 0, len(matches))
	for _, match := range matches {
		fm, ok := match.(*result.FileMatch)
		if !ok {
			return nil, errors.Errorf("expected only file match results")
		}
		fms = append(fms, fm)
	}
	return fms, nil
}

var MockReposContainingPath func() ([]*result.FileMatch, error)

func reposContainingPath(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions, pattern string) ([]*result.FileMatch, error) {
	if MockReposContainingPath != nil {
		return MockReposContainingPath()
	}
	// Use a max FileMatchLimit to ensure we get all the repo matches we
	// can. Setting it to len(repos) could mean we miss some repos since
	// there could be for example len(repos) file matches in the first repo
	// and some more in other repos. deduplicate repo results
	p := search.TextPatternInfo{
		IsRegExp:                     true,
		FileMatchLimit:               math.MaxInt32,
		IncludePatterns:              []string{pattern},
		PathPatternsAreCaseSensitive: false,
		PatternMatchesContent:        true,
		PatternMatchesPath:           true,
	}
	q, err := query.ParseLiteral("file:" + pattern)
	if err != nil {
		return nil, err
	}
	newArgs := *args
	newArgs.PatternInfo = &p
	newArgs.Repos = repos
	newArgs.Query = q
	newArgs.UseFullDeadline = true

	indexed, unindexed, err := zoektutil.PartitionRepos(
		ctx,
		newArgs.Repos,
		newArgs.Zoekt,
		search.TextRequest,
		newArgs.PatternInfo.Index,
		query.ContainsRefGlobs(newArgs.Query),
		func([]*search.RepositoryRevisions) {},
	)
	if err != nil {
		return nil, err
	}

	searcherArgs := &search.SearcherParameters{
		SearcherURLs:    newArgs.SearcherURLs,
		PatternInfo:     newArgs.PatternInfo,
		UseFullDeadline: newArgs.UseFullDeadline,
	}

	agg := streaming.NewAggregatingStream()

	g, ctx := errgroup.WithContext(ctx)

	if newArgs.Mode != search.SearcherOnly {
		typ := search.TextRequest
		zoektQuery, err := search.QueryToZoektQuery(newArgs.PatternInfo, &newArgs.Features, typ)
		if err != nil {
			return nil, err
		}

		zoektArgs := search.ZoektParameters{
			Query:          zoektQuery,
			Typ:            typ,
			FileMatchLimit: newArgs.PatternInfo.FileMatchLimit,
			Select:         newArgs.PatternInfo.Select,
			Zoekt:          newArgs.Zoekt,
		}

		zoektJob := &zoektutil.ZoektRepoSubsetSearch{
			Repos:          indexed,
			Query:          zoektArgs.Query,
			Typ:            search.TextRequest,
			FileMatchLimit: zoektArgs.FileMatchLimit,
			Select:         zoektArgs.Select,
			Zoekt:          zoektArgs.Zoekt,
			Since:          nil,
		}

		// Run literal and regexp searches on indexed repositories.
		g.Go(func() error {
			_, err := zoektJob.Run(ctx, nil, agg)
			return err
		})
	}

	// Concurrently run searcher for all unindexed repos regardless whether text or regexp.
	g.Go(func() error {
		searcherJob := &searcher.Searcher{
			PatternInfo:     searcherArgs.PatternInfo,
			Repos:           unindexed,
			Indexed:         false,
			SearcherURLs:    searcherArgs.SearcherURLs,
			UseFullDeadline: searcherArgs.UseFullDeadline,
		}

		_, err := searcherJob.Run(ctx, nil, agg)
		return err
	})

	err = g.Wait()

	matches, matchesErr := matchesToFileMatches(agg.Results)
	if matchesErr != nil && err == nil {
		err = errors.Wrap(matchesErr, "reposContainingPath failed to convert results")
	}

	if err != nil {
		return nil, err
	}
	return matches, nil
}

// reposToAdd determines which repositories should be included in the result set based on whether they fit in the subset
// of repostiories specified in the query's `repohasfile` and `-repohasfile` fields if they exist.
func reposToAdd(ctx context.Context, args *search.TextParameters, repos []*search.RepositoryRevisions) ([]*search.RepositoryRevisions, error) {
	// matchCounts will contain the count of repohasfile patterns that matched.
	// For negations, we will explicitly set this to -1 if it matches.
	matchCounts := make(map[api.RepoID]int)
	if len(args.PatternInfo.FilePatternsReposMustInclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustInclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}

			matchedIDs := make(map[api.RepoID]struct{})
			for _, m := range matches {
				matchedIDs[m.Repo.ID] = struct{}{}
			}

			// increment the count for all seen repos
			for id := range matchedIDs {
				matchCounts[id] += 1
			}
		}
	} else {
		// Default to including all the repos, then excluding some of them below.
		for _, r := range repos {
			matchCounts[r.Repo.ID] = 0
		}
	}

	if len(args.PatternInfo.FilePatternsReposMustExclude) > 0 {
		for _, pattern := range args.PatternInfo.FilePatternsReposMustExclude {
			matches, err := reposContainingPath(ctx, args, repos, pattern)
			if err != nil {
				return nil, err
			}
			for _, m := range matches {
				matchCounts[m.Repo.ID] = -1
			}
		}
	}

	var rsta []*search.RepositoryRevisions
	for _, r := range repos {
		if count, ok := matchCounts[r.Repo.ID]; ok && count == len(args.PatternInfo.FilePatternsReposMustInclude) {
			rsta = append(rsta, r)
		}
	}

	return rsta, nil
}
