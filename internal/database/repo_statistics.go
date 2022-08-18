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
	// Start transaction
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	// Get current state and the first row we have in the database
	var (
		rowID int
		rs    RepoStatistics
	)
	row := tx.QueryRow(ctx, sqlf.Sprintf(getRepoStatisticsForCompactionQueryFmtstr))
	err = row.Scan(&rowID, &rs.Total, &rs.SoftDeleted, &rs.NotCloned, &rs.Cloning, &rs.Cloned, &rs.FailedFetch)
	if err != nil {
		return err
	}

	// Update the first row so its columns reflect the total state
	updateQuery := sqlf.Sprintf(
		updateRepoStatisticsForCompactionQueryFmtstr,
		rs.Total,
		rs.SoftDeleted,
		rs.NotCloned,
		rs.Cloning,
		rs.Cloned,
		rs.FailedFetch,
		rowID,
	)
	if err := tx.Exec(ctx, updateQuery); err != nil {
		return err
	}

	// Delete all the other rows
	return tx.Exec(ctx, sqlf.Sprintf(deleteOtherRepoStatisticsForCompactionQueryFmtstr, rowID))
}

const getRepoStatisticsForCompactionQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.CompactRepoStatistics
SELECT
	MIN(id), -- choose min id
	SUM(total),
	SUM(soft_deleted),
	SUM(not_cloned),
	SUM(cloning),
	SUM(cloned),
	SUM(failed_fetch)
FROM repo_statistics
`

const updateRepoStatisticsForCompactionQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.CompactRepoStatistics
UPDATE repo_statistics
SET
	total = %s,
	soft_deleted = %s,
	not_cloned = %s,
	cloning = %s,
	cloned = %s,
	failed_fetch = %s
WHERE id = %s;
`

const deleteOtherRepoStatisticsForCompactionQueryFmtstr = `
-- source: internal/database/repo_statistics.go:repoStatisticsStore.CompactRepoStatistics
DELETE FROM repo_statistics WHERE id != %s;
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
