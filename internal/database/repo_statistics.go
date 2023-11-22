package database

import (
	"context"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

// RepoStatistics represents the contents of the single row in the
// repo_statistics table.
type RepoStatistics struct {
	Total       int
	SoftDeleted int
	NotCloned   int
	Cloning     int
	Cloned      int
	FailedFetch int
	Corrupted   int
}

// gitserverRepoStatistics represents the contents of the
// gitserver_repo_statistics table, where each gitserver shard should have a
// separate row and gitserver_repos that haven't been assigned a shard yet have an empty ShardID.
type GitserverReposStatistic struct {
	ShardID     string
	Total       int
	NotCloned   int
	Cloning     int
	Cloned      int
	FailedFetch int
	Corrupted   int
}

type RepoStatisticsStore interface {
	basestore.ShareableStore
	WithTransact(context.Context, func(RepoStatisticsStore) error) error
	With(basestore.ShareableStore) RepoStatisticsStore

	GetRepoStatistics(ctx context.Context) (RepoStatistics, error)
	CompactRepoStatistics(ctx context.Context) error
	GetGitserverReposStatistics(ctx context.Context) ([]GitserverReposStatistic, error)
	CompactGitserverReposStatistics(ctx context.Context) error
}

// repoStatisticsStore is responsible for data stored in the repo_statistics
// and the gitserver_repos_statistics tables.
type repoStatisticsStore struct {
	*basestore.Store
}

// RepoStatisticsWith instantiates and returns a new repoStatisticsStore using
// the other store handle.
func RepoStatisticsWith(other basestore.ShareableStore) RepoStatisticsStore {
	return &repoStatisticsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoStatisticsStore) With(other basestore.ShareableStore) RepoStatisticsStore {
	return &repoStatisticsStore{Store: s.Store.With(other)}
}

func (s *repoStatisticsStore) WithTransact(ctx context.Context, f func(RepoStatisticsStore) error) error {
	return s.Store.WithTransact(ctx, func(tx *basestore.Store) error {
		return f(&repoStatisticsStore{Store: tx})
	})
}

func (s *repoStatisticsStore) GetRepoStatistics(ctx context.Context) (RepoStatistics, error) {
	var rs RepoStatistics
	row := s.QueryRow(ctx, sqlf.Sprintf(getRepoStatisticsQueryFmtstr))
	err := row.Scan(&rs.Total, &rs.SoftDeleted, &rs.NotCloned, &rs.Cloning, &rs.Cloned, &rs.FailedFetch, &rs.Corrupted)
	if err != nil {
		return rs, err
	}
	return rs, nil
}

const getRepoStatisticsQueryFmtstr = `
SELECT
	SUM(total),
	SUM(soft_deleted),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch),
	SUM(corrupted)
FROM repo_statistics
`

func (s *repoStatisticsStore) CompactRepoStatistics(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(compactRepoStatisticsQueryFmtstr))
}

const compactRepoStatisticsQueryFmtstr = `
WITH deleted AS (
	DELETE FROM repo_statistics
	RETURNING
		total,
		soft_deleted,
		not_cloned,
		cloning,
		cloned,
		failed_fetch,
		corrupted
)
INSERT INTO repo_statistics (total, soft_deleted, not_cloned, cloning, cloned, failed_fetch, corrupted)
SELECT
	SUM(total),
	SUM(soft_deleted),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch),
	SUM(corrupted)
FROM deleted;
`

func (s *repoStatisticsStore) CompactGitserverReposStatistics(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(compactGitserverReposStatisticsQueryFmtstr))
}

const compactGitserverReposStatisticsQueryFmtstr = `
WITH deleted AS (
	DELETE FROM gitserver_repos_statistics
	RETURNING
		shard_id,
		total,
		not_cloned,
		cloning,
		cloned,
		failed_fetch,
		corrupted
)
INSERT INTO gitserver_repos_statistics (shard_id, total, not_cloned, cloning, cloned, failed_fetch, corrupted)
SELECT
	shard_id,
	SUM(total),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch),
	SUM(corrupted)
FROM deleted
GROUP BY shard_id;
`

func (s *repoStatisticsStore) GetGitserverReposStatistics(ctx context.Context) ([]GitserverReposStatistic, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(getGitserverReposStatisticsQueryFmtStr))
	return scanGitserverReposStatistics(rows, err)
}

const getGitserverReposStatisticsQueryFmtStr = `
SELECT
	shard_id,
	SUM(total) AS total,
	SUM(not_cloned) AS not_cloned,
	SUM(cloning) AS cloning,
	SUM(cloned) AS cloned,
	SUM(failed_fetch) AS failed_fetch,
	SUM(corrupted) AS corrupted
FROM gitserver_repos_statistics
GROUP BY shard_id;
`

var scanGitserverReposStatistics = basestore.NewSliceScanner(scanGitserverReposStatistic)

func scanGitserverReposStatistic(s dbutil.Scanner) (GitserverReposStatistic, error) {
	var gs = GitserverReposStatistic{}
	err := s.Scan(&gs.ShardID, &gs.Total, &gs.NotCloned, &gs.Cloning, &gs.Cloned, &gs.FailedFetch, &gs.Corrupted)
	if err != nil {
		return gs, err
	}
	return gs, nil
}
