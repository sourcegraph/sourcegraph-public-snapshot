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

type AssignedTeamsStore interface {
	Insert(ctx context.Context, assignedTeamID int32, repoID api.RepoID, absolutePath string, whoAssignedUserID int32) error
	ListAssignedTeamsForRepo(ctx context.Context, repoID api.RepoID) ([]*AssignedTeamSummary, error)
	DeleteOwnerTeam(ctx context.Context, assignedTeamID int32, repoID api.RepoID, absolutePath string) error
}

type AssignedTeamSummary struct {
	OwnerTeamID       int32
	FilePath          string
	RepoID            api.RepoID
	WhoAssignedUserID int32
	AssignedAt        time.Time
}

func AssignedTeamsStoreWith(other basestore.ShareableStore, logger log.Logger) AssignedTeamsStore {
	lgr := logger.Scoped("AssignedTeamsStore")
	return &assignedTeamsStore{Store: basestore.NewWithHandle(other.Handle()), Logger: lgr}
}

type assignedTeamsStore struct {
	*basestore.Store
	Logger log.Logger
}

const insertAssignedTeamFmtstr = `
	WITH repo_path AS (
		SELECT id
		FROM repo_paths
		WHERE absolute_path = %s AND repo_id = %s
	)
	INSERT INTO assigned_teams(owner_team_id, file_path_id, who_assigned_team_id)
	SELECT %s, p.id, %s
	FROM repo_path AS p
	ON CONFLICT DO NOTHING
`

// Insert not only inserts a new assigned team with provided team ID, repo ID
// and path, but it ensures that such a path exists in the first place, after
// which the assigned team is inserted.
func (s assignedTeamsStore) Insert(ctx context.Context, assignedTeamID int32, repoID api.RepoID, absolutePath string, whoAssignedUserID int32) error {
	_, err := ensureRepoPaths(ctx, s.Store, []string{absolutePath}, repoID)
	if err != nil {
		return errors.New("cannot insert repo paths")
	}
	q := sqlf.Sprintf(insertAssignedTeamFmtstr, absolutePath, repoID, assignedTeamID, whoAssignedUserID)
	_, err = s.ExecResult(ctx, q)
	return err
}

const ListAssignedTeamsForRepoFmtstr = `
	SELECT a.owner_team_id, p.absolute_path, p.repo_id, a.who_assigned_team_id, a.assigned_at
	FROM assigned_teams AS a
	INNER JOIN repo_paths AS p ON p.id = a.file_path_id
	WHERE p.repo_id = %s
`

func (s assignedTeamsStore) ListAssignedTeamsForRepo(ctx context.Context, repoID api.RepoID) ([]*AssignedTeamSummary, error) {
	q := sqlf.Sprintf(ListAssignedTeamsForRepoFmtstr, repoID)
	return scanAssignedTeams(s.Query(ctx, q))
}

const deleteAssignedTeamFmtstr = `
	WITH repo_path AS (
		SELECT id
		FROM repo_paths
		WHERE absolute_path = %s AND repo_id = %s
	)
	DELETE FROM assigned_teams
	WHERE owner_team_id = %s AND file_path_id = (
		SELECT p.id
		FROM repo_path AS p
	)
`

func (s assignedTeamsStore) DeleteOwnerTeam(ctx context.Context, assignedTeamID int32, repoID api.RepoID, absolutePath string) error {
	q := sqlf.Sprintf(deleteAssignedTeamFmtstr, absolutePath, repoID, assignedTeamID)
	result, err := s.ExecResult(ctx, q)
	if err != nil {
		return errors.Wrap(err, "executing SQL query")
	}
	deletedRows, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "getting rows affected")
	}
	if deletedRows == 0 {
		return notFoundError{errors.Newf("cannot delete assigned owner team with ID=%d for %q path for repo with ID=%d", assignedTeamID, absolutePath, repoID)}
	}
	return err
}

var scanAssignedTeams = basestore.NewSliceScanner(func(scanner dbutil.Scanner) (*AssignedTeamSummary, error) {
	var summary AssignedTeamSummary
	err := scanAssignedTeam(scanner, &summary)
	return &summary, err
})

func scanAssignedTeam(sc dbutil.Scanner, summary *AssignedTeamSummary) error {
	return sc.Scan(
		&summary.OwnerTeamID,
		&summary.FilePath,
		&summary.RepoID,
		&dbutil.NullInt32{N: &summary.WhoAssignedUserID},
		&summary.AssignedAt,
	)
}
