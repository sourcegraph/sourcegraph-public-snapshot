package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
)

// repoStatistics represents the contents of the single row in the
// repo_statistics table.
type repoStatistics struct {
	Total       int
	SoftDeleted int
	NotCloned   int
	Cloning     int
	Cloned      int
}

// gitserverRepoStatistics represents the contents of the
// gitserver_repo_statistics table, where each gitserver shard should have a
// separate row and gitserver_repos that haven't been assigned a shard yet have an empty ShardID.
type gitserverReposStatistics struct {
	ShardID   string
	Total     int
	NotCloned int
	Cloning   int
	Cloned    int
}

// repoStatisticsStore is responsible for data stored in the repo_statistics
// and the gitserver_repos_statistics tables.
type repoStatisticsStore struct {
	*basestore.Store
}

// RepoStatisticsWith instantiates and returns a new repoStatisticsStore using
// the other store handle.
func RepoStatisticsWith(other basestore.ShareableStore) *repoStatisticsStore {
	return &repoStatisticsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *repoStatisticsStore) With(other basestore.ShareableStore) *repoStatisticsStore {
	return &repoStatisticsStore{Store: s.Store.With(other)}
}

func (s *repoStatisticsStore) Transact(ctx context.Context) (*repoStatisticsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &repoStatisticsStore{Store: txBase}, err
}

func (s *repoStatisticsStore) GetRepoStatistics(ctx context.Context) (repoStatistics, error) {
	var rs repoStatistics
	row := s.QueryRow(ctx, sqlf.Sprintf(getRepoStatisticsQueryFmtstr))
	err := row.Scan(&rs.Total, &rs.SoftDeleted, &rs.NotCloned, &rs.Cloning, &rs.Cloned)
	if err != nil {
		return rs, err
	}
	return rs, nil
}

const getRepoStatisticsQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.GetRepoStatistics
SELECT
	total,
	soft_deleted,
	not_cloned,
	cloning,
	cloned
FROM repo_statistics
`

func (s *repoStatisticsStore) GetGitserverReposStatistics(ctx context.Context) ([]gitserverReposStatistics, error) {
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
	cloned
FROM gitserver_repos_statistics
`

func scanGitserverReposStatistics(rows *sql.Rows) ([]gitserverReposStatistics, error) {
	var out []gitserverReposStatistics
	for rows.Next() {
		gs := gitserverReposStatistics{}
		err := rows.Scan(
			&gs.ShardID,
			&gs.Total,
			&gs.NotCloned,
			&gs.Cloning,
			&gs.Cloned,
		)
		if err != nil {
			return nil, err
		}
		out = append(out, gs)
	}
	return out, nil
}
