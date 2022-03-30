package run

import (
	"context"
	"math"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	"github.com/sourcegraph/sourcegraph/internal/search/result"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

var MockReposContainingPath func() ([]*result.FileMatch, error)

func (s *RepoSearch) reposContainingPath(ctx context.Context, repos []*search.RepositoryRevisions, pattern string) ([]*result.FileMatch, error) {
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
	newArgs := &RepoSearch{
		Query:           q,
		PatternInfo:     &p,
		Repos:           repos,
		RepoOptions:     s.RepoOptions,
		Features:        s.Features,
		Mode:            s.Mode,
		Zoekt:           s.Zoekt,
		SearcherURLs:    s.SearcherURLs,
		UseFullDeadline: true,
	}

	indexed, unindexed, err := zoektutil.PartitionRepos(
		ctx,
		newArgs.Repos,
		newArgs.Zoekt,
		search.TextRequest,
		newArgs.PatternInfo.Index,
		query.ContainsRefGlobs(newArgs.Query),
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

		b, err := query.ToBasicQuery(q)
		if err != nil {
			return nil, err
		}

		types, _ := q.StringValues(query.FieldType)
		var resultTypes result.Types
		if len(types) == 0 {
			resultTypes = result.TypeFile | result.TypePath | result.TypeRepo
		} else {
			for _, t := range types {
				resultTypes = resultTypes.With(result.TypeFromString[t])
			}
		}
		zoektQuery, err := search.QueryToZoektQuery(b, resultTypes, &newArgs.Features, typ)
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
func (s *RepoSearch) reposToAdd(ctx context.Context, repos []*search.RepositoryRevisions) ([]*search.RepositoryRevisions, error) {
	// matchCounts will contain the count of repohasfile patterns that matched.
	// For negations, we will explicitly set this to -1 if it matches.
	matchCounts := make(map[api.RepoID]int)
	if len(s.PatternInfo.FilePatternsReposMustInclude) > 0 {
		for _, pattern := range s.PatternInfo.FilePatternsReposMustInclude {
			matches, err := s.reposContainingPath(ctx, repos, pattern)
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

	if len(s.PatternInfo.FilePatternsReposMustExclude) > 0 {
		for _, pattern := range s.PatternInfo.FilePatternsReposMustExclude {
			matches, err := s.reposContainingPath(ctx, repos, pattern)
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
		if count, ok := matchCounts[r.Repo.ID]; ok && count == len(s.PatternInfo.FilePatternsReposMustInclude) {
			rsta = append(rsta, r)
		}
	}

	return rsta, nil
}
