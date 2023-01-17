package graphqlbackend

import (
	"context"
	"sync"

	"github.com/sourcegraph/sourcegraph/internal/auth"
	"github.com/sourcegraph/sourcegraph/internal/database"
	"github.com/sourcegraph/sourcegraph/internal/gitserver"
	"github.com/sourcegraph/sourcegraph/internal/search"
)

type repositoryStatsResolver struct {
	db database.DB

	indexedStatsOnce  sync.Once
	indexedRepos      int32
	indexedLinesCount int64
	indexedStatsErr   error

	repoStatisticsOnce sync.Once
	repoStatistics     database.RepoStatistics
	repoStatisticsErr  error

	gitDirBytesOnce sync.Once
	gitDirBytes     int64
	gitDirBytesErr  error
}

func (r *repositoryStatsResolver) GitDirBytes(ctx context.Context) (BigInt, error) {
	gitDirBytes, err := r.computeGitDirBytes(ctx)
	if err != nil {
		return 0, err
	}
	return BigInt(gitDirBytes), nil

}

func (r *repositoryStatsResolver) computeGitDirBytes(ctx context.Context) (int64, error) {
	r.gitDirBytesOnce.Do(func() {
		stats, err := gitserver.NewClient(r.db).ReposStats(ctx)
		if err != nil {
			r.gitDirBytesErr = err
			return
		}

		var gitDirBytes int64
		for _, stat := range stats {
			gitDirBytes += stat.GitDirBytes
		}
		r.gitDirBytes = gitDirBytes
	})

	return r.gitDirBytes, r.gitDirBytesErr
}

func (r *repositoryStatsResolver) Indexed(ctx context.Context) (int32, error) {
	indexedRepos, _, err := r.computeIndexedStats(ctx)
	if err != nil {
		return 0, err
	}

	// Since the number of indexed repositories might lag behind the number of
	// repositories in our database (if we recently deleted a repository but
	// Zoekt hasn't removed it from memory yet), we use min(indexed, total)
	// here, so we don't confuse users by returning indexed > total.
	total, err := r.Total(ctx)
	if err != nil {
		return 0, err
	}
	return min(indexedRepos, total), nil
}

func min(a, b int32) int32 {
	if a < b {
		return a
	}
	return b
}

func (r *repositoryStatsResolver) IndexedLinesCount(ctx context.Context) (BigInt, error) {
	_, indexedLinesCount, err := r.computeIndexedStats(ctx)
	if err != nil {
		return 0, err
	}
	return BigInt(indexedLinesCount), nil
}

func (r *repositoryStatsResolver) computeIndexedStats(ctx context.Context) (int32, int64, error) {
	r.indexedStatsOnce.Do(func() {
		repos, err := search.ListAllIndexed(ctx)
		if err != nil {
			r.indexedStatsErr = err
			return
		}
		r.indexedRepos = int32(len(repos.Minimal)) //nolint:staticcheck // See https://github.com/sourcegraph/sourcegraph/issues/45814
		r.indexedLinesCount = int64(repos.Stats.DefaultBranchNewLinesCount) + int64(repos.Stats.OtherBranchesNewLinesCount)
	})

	return r.indexedRepos, r.indexedLinesCount, r.indexedStatsErr
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

func (r *repositoryStatsResolver) Corrupted(ctx context.Context) (int32, error) {
	counts, err := r.computeRepoStatistics(ctx)
	if err != nil {
		return 0, err
	}
	return int32(counts.Corrupted), nil
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
	if err := auth.CheckCurrentUserIsSiteAdmin(ctx, db); err != nil {
		return nil, err
	}

	return &repositoryStatsResolver{db: db}, nil
}
