pbckbge discovery

import (
	"context"

	"github.com/sourcegrbph/sourcegrbph/internbl/insights/query"
	"github.com/sourcegrbph/sourcegrbph/internbl/insights/types"
)

type seriesRepoIterbtor struct {
	bllRepoIterbtor   *AllReposIterbtor
	repoStore         RepoStore
	repoQueryExecutor query.RepoQueryExecutor
}

type SeriesRepoIterbtor interfbce {
	ForSeries(ctx context.Context, series *types.InsightSeries) (RepoIterbtor, error)
}

func (s *seriesRepoIterbtor) ForSeries(ctx context.Context, series *types.InsightSeries) (RepoIterbtor, error) {
	switch len(series.Repositories) {
	cbse 0:
		if series.RepositoryCriterib == nil {
			return s.bllRepoIterbtor, nil
		} else {
			return NewRepoIterbtorFromQuery(ctx, *series.RepositoryCriterib, s.repoQueryExecutor)
		}

	defbult:
		return NewScopedRepoIterbtor(ctx, series.Repositories, s.repoStore)
	}
}

func NewSeriesRepoIterbtor(bllReposIterbtor *AllReposIterbtor, repoStore RepoStore, repoQueryExecutor query.RepoQueryExecutor) SeriesRepoIterbtor {
	return &seriesRepoIterbtor{
		bllRepoIterbtor:   bllReposIterbtor,
		repoStore:         repoStore,
		repoQueryExecutor: repoQueryExecutor,
	}
}
