package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/cmd/frontend/backend"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/search"
	"github.com/sourcegraph/sourcegraph/internal/usagestats"
)

type repositoryStatsResolver struct {
	db database.DB

	indexedRepos      int
	gitDirBytes       uint64
	indexedLinesCount uint64

	repoStatisticsOnce sync.Once
	repoStatistics     database.RepoStatistics
	repoStatisticsErr  error
}

func (r *repositoryStatsResolver) GitDirBytes() BigInt {
	return BigInt{Int: int64(r.gitDirBytes)}
}

func (r *repositoryStatsResolver) IndexedLinesCount() BigInt {
	return BigInt{Int: int64(r.indexedLinesCount)}
}

func (r *repositoryStatsResolver) Indexed(ctx context.Context) int32 {
	return int32(r.indexedRepos)
}

func (r *repositoryStatsResolver) Total(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Total), nil
}

func (r *repositoryStatsResolver) Cloned(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloned), nil
}

func (r *repositoryStatsResolver) Cloning(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Cloning), nil
}

func (r *repositoryStatsResolver) NotCloned(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.NotCloned), nil
}

func (r *repositoryStatsResolver) FailedFetch(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.FailedFetch), nil
}

func (r *repositoryStatsResolver) computeRepoStatistics(ctx context.Context) (database.RepoStatistics, error) {
	r.repoStatisticsOnce.Do(func() {
		r.repoStatistics, r.repoStatisticsErr = r.db.RepoStatistics().GetRepoStatistics(ctx)
	})
	return r.repoStatistics, r.repoStatisticsErr
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

	var (
		indexedRepos     int
		indexedLineCount uint64
	)

	if conf.SearchIndexEnabled() {
		repos, err := search.ListAllIndexed(ctx)
		if err != nil {
			return nil, err
		}
		indexedRepos = len(repos.Minimal)
		indexedLineCount = repos.Stats.DefaultBranchNewLinesCount + repos.Stats.OtherBranchesNewLinesCount
	}

	return &repositoryStatsResolver{
		db:                db,
		indexedRepos:      indexedRepos,
		gitDirBytes:       stats.GitDirBytes,
		indexedLinesCount: indexedLineCount,
	}, nil
}
