package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type AssignedOwnersStore interface {
	Insert(ctx context.Context, assignedOwnerID int32, repoID api.RepoID, absolutePath string, whoAssignedUserID int32) error
	ListAssignedOwnersForRepo(ctx context.Context, repoID api.RepoID) ([]*AssignedOwnerSummary, error)
}

type AssignedOwnerSummary struct {
	OwnerUserID       int32
	FilePath          string
	RepoID            api.RepoID
	WhoAssignedUserID int32
	AssignedAt        time.Time
}

func AssignedOwnersStoreWith(other basestore.ShareableStore, logger log.Logger) AssignedOwnersStore {
	lgr := logger.Scoped("AssignedOwnersStore", "Store for a table containing manually assigned code owners")
	return &assignedOwnersStore{Store: basestore.NewWithHandle(other.Handle()), Logger: lgr}
}

type assignedOwnersStore struct {
	*basestore.Store
	Logger log.Logger
}

const insertAssignedOwnerFmtstr = `
	WITH repo_path AS (
		SELECT id
		FROM repo_paths
		WHERE absolute_path = %s AND repo_id = %s
	)
	INSERT INTO assigned_owners(owner_user_id, file_path_id, who_assigned_user_id)
	SELECT %s, p.id, %s
	FROM repo_path AS p
	ON CONFLICT DO NOTHING
`

func (s assignedOwnersStore) Insert(ctx context.Context, assignedOwnerID int32, repoID api.RepoID, absolutePath string, whoAssignedUserID int32) error {
	q := sqlf.Sprintf(insertAssignedOwnerFmtstr, absolutePath, repoID, assignedOwnerID, whoAssignedUserID)
	result, err := s.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "executing SQL query")
	}
	insertedRows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting rows affected")
	}
	if insertedRows == 0 {
		return notFoundError{errors.Newf("cannot find %q path for repo with ID=%d", absolutePath, repoID)}
	}
	return err
}

const listAssignedOwnersForRepoFmtstr = `
	SELECT a.owner_user_id, p.absolute_path, p.repo_id, a.who_assigned_user_id, a.assigned_at
	FROM assigned_owners AS a
	INNER JOIN repo_paths AS p ON p.id = a.file_path_id
	WHERE p.repo_id = %s
`

func (s assignedOwnersStore) ListAssignedOwnersForRepo(ctx context.Context, repoID api.RepoID) ([]*AssignedOwnerSummary, error) {
	q := sqlf.Sprintf(listAssignedOwnersForRepoFmtstr, repoID)
	return scanAssignedOwners(s.Query(ctx, q))
}

var scanAssignedOwners = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (*AssignedOwnerSummary, error) {
	var summary AssignedOwnerSummary
	err := scanAssignedOwner(scanner, &summary)
	return &summary, err
})

func scanAssignedOwner(sc dbutil.Scanner, summary *AssignedOwnerSummary) error {
	return sc.Scan(
		&summary.OwnerUserID,
		&summary.FilePath,
		&summary.RepoID,
		&dbutil.NullInt32{N: &summary.WhoAssignedUserID},
		&summary.AssignedAt,
	)
}
