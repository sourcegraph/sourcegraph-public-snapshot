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

		indexed, unindexed, err := zoektutil.PartitionRepos(
			ctx,
			page.RepoRevs,
			t.ZoektArgs.Zoekt,
			search.TextRequest,
			t.UseIndex,
			t.ContainsRefGlobs,
			zoektutil.MissingRepoRevStatus(stream),
		)
		if err != nil {
			return err
		}

		g, ctx := errgroup.WithContext(ctx)

		if t.NotSearcherOnly {
			zoektJob := &zoektutil.ZoektRepoSubsetSearch{
				Repos:          indexed,
				Query:          t.ZoektArgs.Query,
				Typ:            search.TextRequest,
				FileMatchLimit: t.ZoektArgs.FileMatchLimit,
				Select:         t.ZoektArgs.Select,
				Zoekt:          t.ZoektArgs.Zoekt,
				Since:          nil,
			}

			// Run literal and regexp searches on indexed repositories.
			g.Go(func() error {
				_, err := zoektJob.Run(ctx, db, stream)
				return err
			})
		}

		// Concurrently run searcher for all unindexed repos regardless whether text or regexp.
		g.Go(func() error {
			searcherJob := &searcher.Searcher{
				PatternInfo:     t.SearcherArgs.PatternInfo,
				Repos:           unindexed,
				Indexed:         false,
				SearcherURLs:    t.SearcherArgs.SearcherURLs,
				UseFullDeadline: t.SearcherArgs.UseFullDeadline,
			}
			_, err := searcherJob.Run(ctx, db, stream)
			return err
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
