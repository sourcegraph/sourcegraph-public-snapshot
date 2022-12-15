package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type seriesRepoIterator struct {
	allRepoIterator *AllReposIterator
	repoStore       RepoStore
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
			return nil, errors.New("search scoped repository series not implemented")
		}

	default:
		return NewScopedRepoIterator(ctx, series.Repositories, s.repoStore)
	}
}

func NewSeriesRepoIterator(allReposIterator *AllReposIterator, repoStore RepoStore) SeriesRepoIterator {
	return &seriesRepoIterator{
		allRepoIterator: allReposIterator,
		repoStore:       repoStore,
	}
}
