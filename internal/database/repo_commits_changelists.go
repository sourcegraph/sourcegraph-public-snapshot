package database

import (
	"context"
	"database/sql"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/perforce"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type RepoCommitsChangelistsStore interface {
	// BatchInsertCommitSHAsWithPerforceChangelistID will insert rows into the
	// repo_commits_changelists table in batches.
	BatchInsertCommitSHAsWithPerforceChangelistID(context.Context, api.RepoID, []types.PerforceChangelist) error
	// GetLatestForRepo will return the latest commit that has been mapped in the database.
	GetLatestForRepo(ctx context.Context, repoID api.RepoID) (*types.RepoCommit, error)

	// GetRepoCommit will return the matching row from the table for the given repo ID and the
	// given changelist ID.
	GetRepoCommitChangelist(ctx context.Context, repoID api.RepoID, changelistID int64) (*types.RepoCommit, error)

	// BatchGetRepoCommitChangelist bulk loads repo commits for given repo ids and changelistIds
	BatchGetRepoCommitChangelist(ctx context.Context, rcs ...RepoChangelistIDs) (map[api.RepoID]map[int64]*types.RepoCommit, error)
}

type repoCommitsChangelistsStore struct {
	*basestore.Store
	logger log.Logger
}

var _ RepoCommitsChangelistsStore = (*repoCommitsChangelistsStore)(nil)

func RepoCommitsChangelistsWith(logger log.Logger, other basestore.ShareableStore) RepoCommitsChangelistsStore {
	return &repoCommitsChangelistsStore{
		logger: logger,
		Store:  basestore.NewWithHandle(other.Handle()),
	}
}

func (s *repoCommitsChangelistsStore) BatchInsertCommitSHAsWithPerforceChangelistID(ctx context.Context, repo_id api.RepoID, commitsMap []types.PerforceChangelist) error {
	return s.WithTransact(ctx, func(tx *basestore.Store) error {

		inserter := batch.NewInserter(ctx, tx.Handle(), "repo_commits_changelists", batch.MaxNumPostgresParameters, "repo_id", "commit_sha", "perforce_changelist_id")
		for _, item := range commitsMap {
			if err := inserter.Insert(
				ctx,
				int32(repo_id),
				dbutil.CommitBytea(item.CommitSHA),
				item.ChangelistID,
			); err != nil {
				return err
			}
		}
		return inserter.Flush(ctx)
	})

}

var getLatestForRepoFmtStr = `
SELECT
	id,
	repo_id,
	commit_sha,
	perforce_changelist_id
	created_at
FROM
	repo_commits_changelists
WHERE
	repo_id = %s
ORDER BY
	perforce_changelist_id DESC
LIMIT 1`

func (s *repoCommitsChangelistsStore) GetLatestForRepo(ctx context.Context, repoID api.RepoID) (*types.RepoCommit, error) {
	q := sqlf.Sprintf(getLatestForRepoFmtStr, repoID)
	row := s.QueryRow(ctx, q)
	return scanRepoCommitRow(row)
}

func scanRepoCommitRow(scanner dbutil.Scanner) (*types.RepoCommit, error) {
	var r types.RepoCommit
	if err := scanner.Scan(
		&r.ID,
		&r.RepoID,
		&r.CommitSHA,
		&r.PerforceChangelistID,
	); err != nil {
		return nil, err
	}

	return &r, nil
}

var getRepoCommitFmtStr = `
SELECT
	id,
	repo_id,
	commit_sha,
	perforce_changelist_id
FROM
	repo_commits_changelists
WHERE
	repo_id = %s
	AND perforce_changelist_id = %s;
`

func (s *repoCommitsChangelistsStore) GetRepoCommitChangelist(ctx context.Context, repoID api.RepoID, changelistID int64) (*types.RepoCommit, error) {
	q := sqlf.Sprintf(getRepoCommitFmtStr, repoID, changelistID)

	repoCommit, err := scanRepoCommitRow(s.QueryRow(ctx, q))
	if err == sql.ErrNoRows {
		return nil, &perforce.ChangelistNotFoundError{RepoID: repoID, ID: changelistID}
	} else if err != nil {
		return nil, err
	}
	return repoCommit, nil
}

type RepoChangelistIDs struct {
	RepoID        api.RepoID
	ChangelistIDs []int64
}

var getRepoCommitFmtBatchStr = `
SELECT
	id,
	repo_id,
	commit_sha,
	perforce_changelist_id
FROM
	repo_commits_changelists
WHERE
	%s;
`

func (s *repoCommitsChangelistsStore) BatchGetRepoCommitChangelist(ctx context.Context, rcs ...RepoChangelistIDs) (map[api.RepoID]map[int64]*types.RepoCommit, error) {
	res := make(map[api.RepoID]map[int64]*types.RepoCommit, len(rcs))
	for _, rc := range rcs {
		res[rc.RepoID] = make(map[int64]*types.RepoCommit, len(rc.ChangelistIDs))
	}

	var where []*sqlf.Query
	for _, rc := range rcs {
		changeListIdsLength := len(rc.ChangelistIDs)
		if changeListIdsLength == 0 {
			continue
		}
		items := make([]*sqlf.Query, changeListIdsLength)
		for i, id := range rc.ChangelistIDs {
			items[i] = sqlf.Sprintf("%d", id)
		}
		where = append(where, sqlf.Sprintf("(repo_id=%d AND perforce_changelist_id IN (%s))", rc.RepoID, sqlf.Join(items, ",")))
	}

	var whereClause *sqlf.Query
	if len(where) > 0 {
		whereClause = sqlf.Join(where, "\n OR ")
	} else {
		// If input has no changeList ids just return an empty result
		return res, nil
	}

	q := sqlf.Sprintf(getRepoCommitFmtBatchStr, whereClause)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if repoCommit, err := scanRepoCommitRow(rows); err != nil {
			return nil, err
		} else {
			res[repoCommit.RepoID][repoCommit.PerforceChangelistID] = repoCommit
		}
	}

	return res, rows.Err()
}
