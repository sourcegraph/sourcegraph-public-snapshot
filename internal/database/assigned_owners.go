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
	DeleteOwner(ctx context.Context, assignedOwnerID int32, repoID api.RepoID, absolutePath string) error
	CountAssignedOwners(ctx context.Context) (int32, error)
}

type AssignedOwnerSummary struct {
	OwnerUserID       int32
	FilePath          string
	RepoID            api.RepoID
	WhoAssignedUserID int32
	AssignedAt        time.Time
}

func AssignedOwnersStoreWith(other basestore.ShareableStore, logger log.Logger) AssignedOwnersStore {
	lgr := logger.Scoped("AssignedOwnersStore")
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

// Insert not only inserts a new assigned owner with provided user ID, repo ID
// and path, but it ensures that such a path exists in the first place, after
// which the user is inserted.
func (s assignedOwnersStore) Insert(ctx context.Context, assignedOwnerID int32, repoID api.RepoID, absolutePath string, whoAssignedUserID int32) error {
	_, err := ensureRepoPaths(ctx, s.Store, []string{absolutePath}, repoID)
	if err != nil {
		return errors.New("cannot insert repo paths")
	}
	q := sqlf.Sprintf(insertAssignedOwnerFmtstr, absolutePath, repoID, assignedOwnerID, whoAssignedUserID)
	_, err = s.ExecResult(ctx, q)
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

const deleteAssignedOwnerFmtstr = `
	WITH repo_path AS (
		SELECT id
		FROM repo_paths
		WHERE absolute_path = %s AND repo_id = %s
	)
	DELETE FROM assigned_owners
	WHERE owner_user_id = %s AND file_path_id = (
		SELECT p.id
		FROM repo_path AS p
	)
`

func (s assignedOwnersStore) DeleteOwner(ctx context.Context, assignedOwnerID int32, repoID api.RepoID, absolutePath string) error {
	q := sqlf.Sprintf(deleteAssignedOwnerFmtstr, absolutePath, repoID, assignedOwnerID)
	result, err := s.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "executing SQL query")
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting rows affected")
	}
	if deletedRows == 0 {
		return notFoundError{errors.Newf("cannot delete assigned owner with ID=%d for %q path for repo with ID=%d", assignedOwnerID, absolutePath, repoID)}
	}
	return err
}

const countAssignedOwnersFmtstr = `SELECT COUNT(*) FROM assigned_owners`

func (s assignedOwnersStore) CountAssignedOwners(ctx context.Context) (int32, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(countAssignedOwnersFmtstr)))
	if err != nil {
		return 0, err
	}
	return int32(count), err
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
