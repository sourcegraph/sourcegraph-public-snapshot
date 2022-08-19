package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

type repositoryStatsResolver struct {
	db database.DB

	gitDirBytes       uint64
	indexedLinesCount uint64

	statisticsCountsOnce sync.Once
	statisticsCounts     database.StatisticsCounts
	statisticsCountsErr  error
}

func (r *repositoryStatsResolver) GitDirBytes() BigInt {
	return BigInt{Int: int64(r.gitDirBytes)}
}

func (r *repositoryStatsResolver) IndexedLinesCount() BigInt {
	return BigInt{Int: int64(r.indexedLinesCount)}
}

func (r *repositoryStatsResolver) Total(ctx context.Context) (int32, error) {
	counts, err := r.computeStatisticsCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Total), nil
}

func (r *repositoryStatsResolver) Cloned(ctx context.Context) (int32, error) {
	counts, err := r.computeStatisticsCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloned), nil
}

func (r *repositoryStatsResolver) Cloning(ctx context.Context) (int32, error) {
	counts, err := r.computeStatisticsCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloning), nil
}

func (r *repositoryStatsResolver) NotCloned(ctx context.Context) (int32, error) {
	counts, err := r.computeStatisticsCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.NotCloned), nil
}

func (r *repositoryStatsResolver) FailedFetch(ctx context.Context) (int32, error) {
	counts, err := r.computeStatisticsCounts(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.FailedFetch), nil
}

func (r *repositoryStatsResolver) computeStatisticsCounts(ctx context.Context) (database.StatisticsCounts, error) {
	r.statisticsCountsOnce.Do(func() {
		r.statisticsCounts, r.statisticsCountsErr = r.db.Repos().StatisticsCounts(ctx)
	})
	return r.statisticsCounts, r.statisticsCountsErr
}

func (r *schemaResolver) RepositoryStats(ctx context.Context) (*repositoryStatsResolver, error) {
	// 🚨 SECURITY: Only site admins may query repository statistics for the site.
	db := r.db
	if err := backend.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	stats, err := usagestats.GetRepositories(ctx, db)
	if err != nil {
		return nil, err
	}

	return &repositoryStatsResolver{
		db:                db,
		gitDirBytes:       stats.GitDirBytes,
		indexedLinesCount: stats.DefaultBranchNewLinesCount + stats.OtherBranchesNewLinesCount,
	}, nil
}
