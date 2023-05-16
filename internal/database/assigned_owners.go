package database

import (
	"context"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/log"
	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
)

type AssignedOwnersStore interface {
	Insert(ctx context.Context, assignedOwnerID int32, filePathID int, whoAssignedUserID int32) error
	ListByFilePath(ctx context.Context, filePath string) ([]*AssignedOwnerSummary, error)
	ListAssignedOwnersForRepo(ctx context.Context, repoID api.RepoID) ([]*AssignedOwnerSummary, error)
}

type AssignedOwnerSummary struct {
	UserID            int32
	FilePathID        int
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
	INSERT INTO assigned_owners(user_id, file_path_id, who_assigned_user_id)
	VALUES (%s, %s, %s)
	ON CONFLICT DO NOTHING
`

func (s assignedOwnersStore) Insert(ctx context.Context, assignedOwnerID int32, filePathID int, whoAssignedUserID int32) error {
	q := sqlf.Sprintf(insertAssignedOwnerFmtstr, assignedOwnerID, filePathID, whoAssignedUserID)
	return s.Exec(ctx, q)
}

const listAssignedOwnersForRepoPathFmtstr = `
	SELECT a.user_id, a.file_path_id, a.who_assigned_user_id, a.assigned_at
	FROM assigned_owners AS a
	INNER JOIN repo_paths AS p ON p.id = a.file_path_id
	WHERE p.absolute_path = %s
`

func (s assignedOwnersStore) ListByFilePath(ctx context.Context, filePath string) ([]*AssignedOwnerSummary, error) {
	q := sqlf.Sprintf(listAssignedOwnersForRepoPathFmtstr, filePath)
	return scanAssignedOwners(s.Query(ctx, q))
}

const listAssignedOwnersForRepoFmtstr = `
	SELECT a.user_id, a.file_path_id, a.who_assigned_user_id, a.assigned_at
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
		&summary.UserID,
		&summary.FilePathID,
		&dbutil.NullInt32{N: &summary.WhoAssignedUserID},
		&summary.AssignedAt,
	)
}
