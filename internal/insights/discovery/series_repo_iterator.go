package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/internal/insights/query"
	"github.com/sourcegraph/sourcegraph/internal/insights/types"
)

type seriesRepoIterator struct {
	allRepoIterator   *AllReposIterator
	repoStore         RepoStore
	repoQueryExecutor query.RepoQueryExecutor
}

type SeriesRepoIterator interface {
	ForSeries(ctx context.Context, series *types.InsightSeries) (RepoIterator, error)
}

func (s *seriesRepoIterator) ForSeries(ctx context.Context, series *types.InsightSeries) (RepoIterator, error) {
	switch len(series.Repositories) {
	case 0:
		if series.RepositoryCriteria == nil {
			return s.allRepoIterator, nil
		} else {
			return NewRepoIteratorFromQuery(ctx, *series.RepositoryCriteria, s.repoQueryExecutor)
		}

	default:
		return NewScopedRepoIterator(ctx, series.Repositories, s.repoStore)
	}
}

func NewSeriesRepoIterator(allReposIterator *AllReposIterator, repoStore RepoStore, repoQueryExecutor query.RepoQueryExecutor) SeriesRepoIterator {
	return &seriesRepoIterator{
		allRepoIterator:   allReposIterator,
		repoStore:         repoStore,
		repoQueryExecutor: repoQueryExecutor,
	}
}
