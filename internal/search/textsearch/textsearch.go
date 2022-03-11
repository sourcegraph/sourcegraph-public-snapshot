package textsearch

import (
	"context"

	"golang.org/x/sync/errgroup"

	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	searchrepos "github.com/sourcegraph/sourcegraph/internal/search/repos"
	"github.com/sourcegraph/sourcegraph/internal/search/streaming"
	zoektutil "github.com/sourcegraph/sourcegraph/internal/search/zoekt"
	"github.com/sourcegraph/sourcegraph/internal/trace"
)

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
