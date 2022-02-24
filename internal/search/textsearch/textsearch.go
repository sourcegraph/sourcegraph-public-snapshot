package textsearch

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/search/query"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/searcher"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

type RepoSubsetTextSearch struct {
	ZoektArgs        *search.ZoektParameters
	SearcherArgs     *search.SearcherParameters
	NotSearcherOnly  bool
	UseIndex         query.YesNoOnly
	ContainsRefGlobs bool

	RepoOpts search.RepoOptions
}

func (t *RepoSubsetTextSearch) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "RepoSubsetTextSearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	repos := &searchrepos.Resolver{DB: db, Opts: t.RepoOpts}
	return nil, repos.Paginate(ctx, nil, func(page *searchrepos.Resolved) error {
		request, ok, err := zoektutil.OnlyUnindexed(page.RepoRevs, t.ZoektArgs.Zoekt, t.UseIndex, t.ContainsRefGlobs, zoektutil.MissingRepoRevStatus(stream))
		if err != nil {
			return err
		}

		if !ok {
			request, err = zoektutil.NewIndexedSubsetSearchRequest(ctx, page.RepoRevs, t.UseIndex, t.ZoektArgs, zoektutil.MissingRepoRevStatus(stream))
			if err != nil {
				return err
			}
		}

		g, ctx := errgroup.WithContext(ctx)

		if t.NotSearcherOnly {
			// Run literal and regexp searches on indexed repositories.
			g.Go(func() error {
				return request.Search(ctx, stream)
			})
		}

		// Concurrently run searcher for all unindexed repos regardless whether text or regexp.
		g.Go(func() error {
			return searcher.SearchOverRepos(ctx, t.SearcherArgs, stream, request.UnindexedRepos(), false)
		})

		return g.Wait()
	})
}

func (*RepoSubsetTextSearch) Name() string {
	return "RepoSubsetText"
}

type RepoUniverseTextSearch struct {
	GlobalZoektQuery *zoektutil.GlobalZoektQuery
	ZoektArgs        *search.ZoektParameters

	RepoOptions search.RepoOptions
	UserID      int32
}

func (t *RepoUniverseTextSearch) Run(ctx context.Context, db database.DB, stream streaming.Sender) (_ *search.Alert, err error) {
	tr, ctx := trace.New(ctx, "RepoUniverseTextSearch", "")
	defer func() {
		tr.SetError(err)
		tr.Finish()
	}()

	userPrivateRepos := searchrepos.PrivateReposForActor(ctx, db, t.RepoOptions)
	t.GlobalZoektQuery.ApplyPrivateFilter(userPrivateRepos)
	t.ZoektArgs.Query = t.GlobalZoektQuery.Generate()

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return zoektutil.DoZoektSearchGlobal(ctx, t.ZoektArgs, stream)
	})
	return nil, g.Wait()
}

func (*RepoUniverseTextSearch) Name() string {
	return "RepoUniverseText"
}
