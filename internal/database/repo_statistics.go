package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
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
}

// gitserverRepoStatistics represents the contents of the
// gitserver_repo_statistics table, where each gitserver shard should have a
// separate row and gitserver_repos that haven't been assigned a shard yet have an empty ShardID.
type GitserverReposStatistics struct {
	ShardID     string
	Total       int
	NotCloned   int
	Cloning     int
	Cloned      int
	FailedFetch int
}

type RepoStatisticsStore interface {
	Transact(context.Context) (RepoStatisticsStore, error)
	With(basestore.ShareableStore) RepoStatisticsStore

	GetRepoStatistics(ctx context.Context) (RepoStatistics, error)
	CompactRepoStatistics(ctx context.Context) error
	GetGitserverReposStatistics(ctx context.Context) ([]GitserverReposStatistics, error)
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

func (s *repoStatisticsStore) Transact(ctx context.Context) (RepoStatisticsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &repoStatisticsStore{Store: txBase}, err
}

func (s *repoStatisticsStore) GetRepoStatistics(ctx context.Context) (RepoStatistics, error) {
	var rs RepoStatistics
	row := s.QueryRow(ctx, sqlf.Sprintf(getRepoStatisticsQueryFmtstr))
	err := row.Scan(&rs.Total, &rs.SoftDeleted, &rs.NotCloned, &rs.Cloning, &rs.Cloned, &rs.FailedFetch)
	if err != nil {
		return rs, err
	}
	return rs, nil
}

const getRepoStatisticsQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.GetRepoStatistics
SELECT
	SUM(total),
	SUM(soft_deleted),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch)
FROM repo_statistics
`

func (s *repoStatisticsStore) CompactRepoStatistics(ctx context.Context) error {
	return s.Exec(ctx, sqlf.Sprintf(compactRepoStatisticsQueryFmtstr))
}

const compactRepoStatisticsQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.CompactRepoStatistics
WITH deleted AS (
	DELETE FROM repo_statistics
	RETURNING
		total,
		soft_deleted,
		not_cloned,
		cloning,
		cloned,
		failed_fetch
)
INSERT INTO repo_statistics (total, soft_deleted, not_cloned, cloning, cloned, failed_fetch)
SELECT
	SUM(total),
	SUM(soft_deleted),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch)
FROM deleted;
`

func (s *repoStatisticsStore) GetGitserverReposStatistics(ctx context.Context) ([]GitserverReposStatistics, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(getGitserverReposStatisticsQueryFmtStr))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return scanGitserverReposStatistics(rows)
}

const getGitserverReposStatisticsQueryFmtStr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.GetGitserverReposStatistics
SELECT
	shard_id,
	total,
	not_cloned,
	cloning,
	cloned,
	failed_fetch
FROM gitserver_repos_statistics
`

func scanGitserverReposStatistics(rows *sql.Rows) ([]GitserverReposStatistics, error) {
	var out []GitserverReposStatistics
	for rows.Next() {
		gs := GitserverReposStatistics{}
		err := rows.Scan(
			&gs.ShardID,
			&gs.Total,
			&gs.NotCloned,
			&gs.Cloning,
			&gs.Cloned,
			&gs.FailedFetch,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, gs)
	}
	return out, nil
}
