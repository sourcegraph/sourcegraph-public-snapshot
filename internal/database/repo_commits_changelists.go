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

	// GetRepoCommit will return the mathcing row from the table for the given repo ID and the
	// given changelist ID.
	GetRepoCommitChangelist(ctx context.Context, repoID api.RepoID, changelistID int64) (*types.RepoCommit, error)
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
