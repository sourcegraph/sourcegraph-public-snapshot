package discovery

import (
	"context"

	"github.com/sourcegraph/sourcegraph/enterprise/internal/insights/types"
)

type seriesRepoIterator struct {
	allRepoIterator *AllReposIterator
	repoStore       RepoStore
}

type SeriesRepoIterator interface {
	SeriesRepoIterator(ctx context.Context, series *types.InsightSeries) (RepoIterator, error)
}

func (s *seriesRepoIterator) SeriesRepoIterator(ctx context.Context, series *types.InsightSeries) (RepoIterator, error) {
	switch len(series.Repositories) {
	case 0:
		return s.allRepoIterator, nil
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
