package database

import (
	"context"
	"database/sql"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

// GitserverReposStore is responsible for data stored in the gitserver_repos table.
type GitserverRepoStore struct {
	*basestore.Store
}

// GitserverRepos instantiates and returns a new GitserverRepoStore.
func GitserverRepos(db dbutil.DB) *GitserverRepoStore {
	return &GitserverRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewGitserverReposWith instantiates and returns a new GitserverRepoStore using
// the other store handle.
func NewGitserverReposWith(other basestore.ShareableStore) *GitserverRepoStore {
	return &GitserverRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *GitserverRepoStore) With(other basestore.ShareableStore) *GitserverRepoStore {
	return &GitserverRepoStore{Store: s.Store.With(other)}
}

func (s *GitserverRepoStore) Transact(ctx context.Context) (*GitserverRepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &GitserverRepoStore{Store: txBase}, err
}

// Upsert adds a row representing the GitServer status of a repo
func (s *GitserverRepoStore) Upsert(ctx context.Context, repos ...*types.GitserverRepo) error {
	values := make([]*sqlf.Query, 0, len(repos))
	for _, gr := range repos {
		q := sqlf.Sprintf("(%s, %s, %s, %s, %s, now())",
			gr.RepoID,
			gr.CloneStatus,
			dbutil.NewNullString(gr.ShardID),
			dbutil.NewNullInt64(gr.LastExternalService),
			dbutil.NewNullString(sanitizeToUTF8(gr.LastError)),
		)

		values = append(values, q)
	}

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:GitserverRepoStore.Upsert
INSERT INTO
    gitserver_repos(repo_id, clone_status, shard_id, last_external_service, last_error, updated_at)
    VALUES %s
    ON CONFLICT (repo_id) DO UPDATE
    SET (clone_status, shard_id, last_external_service, last_error, updated_at) =
        (EXCLUDED.clone_status, EXCLUDED.shard_id, EXCLUDED.last_external_service, EXCLUDED.last_error, now())
`, sqlf.Join(values, ",")))

	return errors.Wrap(err, "creating GitserverRepo")
}

type IterateRepoGitserverStatusOptions struct {
	// If set, will only iterate over repos that have not been assigned to a shard
	OnlyWithoutShard bool
}

// IterateRepoGitserverStatus iterates over the status of all repos by joining
// our repo and gitserver_repos table. It is possible for us not to have a
// corresponding row in gitserver_repos yet. repoFn will be called once for each
// row. If it returns an error we'll abort iteration.
func (s *GitserverRepoStore) IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error {
	if repoFn == nil {
		return errors.New("nil repoFn")
	}

	q := `
-- source: internal/database/gitserver_repos.go:GitserverRepoStore.IterateRepoGitserverStatus
SELECT repo.id,
       repo.name,
       gr.clone_status,
       gr.shard_id,
       gr.last_external_service,
       gr.last_error,
       gr.updated_at
FROM repo
    LEFT JOIN gitserver_repos gr ON gr.repo_id = repo.id
    WHERE repo.deleted_at IS NULL
`
	if options.OnlyWithoutShard {
		q = q + "AND (gr.shard_id = '' OR gr IS NULL)"
	}

	rows, err := s.Query(ctx, sqlf.Sprintf(q))
	if err != nil {
		return errors.Wrap(err, "fetching gitserver status")
	}
	defer rows.Close()

	for rows.Next() {
		var rgs types.RepoGitserverStatus
		var gr types.GitserverRepo
		var cloneStatus string

		if err := rows.Scan(
			&rgs.ID,
			&rgs.Name,
			&dbutil.NullString{S: &cloneStatus},
			&dbutil.NullString{S: &gr.ShardID},
			&dbutil.NullInt64{N: &gr.LastExternalService},
			&dbutil.NullString{S: &gr.LastError},
			&dbutil.NullTime{Time: &gr.UpdatedAt},
		); err != nil {
			return errors.Wrap(err, "scanning row")
		}

		// Clone status will only be null if we don't have a corresponding row in
		// gitserver_repos
		if cloneStatus != "" {
			gr.CloneStatus = types.ParseCloneStatus(cloneStatus)
			gr.RepoID = rgs.ID
			rgs.GitserverRepo = &gr
		}

		err := repoFn(rgs)
		if err != nil {
			// Abort
			return errors.Wrap(err, "calling repoFn")
		}
	}

	if rows.Err() != nil {
		return errors.Wrap(rows.Err(), "iterating rows")
	}

	return nil
}

func (s *GitserverRepoStore) GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error) {
	q := `
-- source: internal/database/gitserver_repos.go:GitserverRepoStore.GetByID
SELECT
       repo_id,
       clone_status,
       shard_id,
       last_external_service,
       last_error,
       updated_at
FROM gitserver_repos
WHERE repo_id = %s
`

	row := s.QueryRow(ctx, sqlf.Sprintf(q, id))
	if row.Err() != nil {
		return nil, errors.Wrap(row.Err(), "getting GitserverRepo")
	}
	var gr types.GitserverRepo
	var cloneStatus string
	err := row.Scan(
		&gr.RepoID,
		&cloneStatus,
		&gr.ShardID,
		&dbutil.NullInt64{N: &gr.LastExternalService},
		&dbutil.NullString{S: &gr.LastError},
		&gr.UpdatedAt,
	)
	if err != nil {
		return nil, errors.Wrap(err, "scanning GitserverRepo")
	}
	gr.CloneStatus = types.ParseCloneStatus(cloneStatus)

	return &gr, nil
}

// SetCloneStatus will attempt to update ONLY the clone status of a
// GitServerRepo. If a matching row does not yet exist a new one will be created.
// If the status value hasn't changed, the row will not be updated.
func (s *GitserverRepoStore) SetCloneStatus(ctx context.Context, id api.RepoID, status types.CloneStatus, shardID string) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:GitserverRepoStore.SetCloneStatus
INSERT INTO gitserver_repos(repo_id, clone_status, shard_id, updated_at)
VALUES (%s, %s, %s, now())
ON CONFLICT (repo_id) DO UPDATE
SET (clone_status, shard_id, updated_at) =
    (EXCLUDED.clone_status, EXCLUDED.shard_id, now())
    WHERE gitserver_repos.clone_status IS DISTINCT FROM EXCLUDED.clone_status
`, id, status, shardID))

	return errors.Wrap(err, "setting clone status")
}

// SetLastError will attempt to update ONLY the last error of a GitServerRepo. If
// a matching row does not yet exist a new one will be created.
// If the error value hasn't changed, the row will not be updated.
func (s *GitserverRepoStore) SetLastError(ctx context.Context, id api.RepoID, error, shardID string) error {
	ns := dbutil.NewNullString(sanitizeToUTF8(error))

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:GitserverRepoStore.SetLastError
INSERT INTO gitserver_repos(repo_id, last_error, shard_id, updated_at)
VALUES (%s, %s, %s, now())
ON CONFLICT (repo_id) DO UPDATE
SET (last_error, shard_id, updated_at) =
    (EXCLUDED.last_error, EXCLUDED.shard_id, now())
    WHERE gitserver_repos.last_error IS DISTINCT FROM EXCLUDED.last_error
`, id, ns, shardID))

	return errors.Wrap(err, "setting last error")
}

// sanitizeToUTF8 will remove any null character terminated string. The null character can be
// represented in one of the following ways in Go:
//
// Hex: \x00
// Unicode: \u0000
// Octal digits: \000
//
// Using any of them to replace the null character has the same effect. See this playground
// example: https://play.golang.org/p/8SKPmalJRRW
//
// See this for a detailed answer:
// https://stackoverflow.com/a/38008565/1773961
func sanitizeToUTF8(s string) string {
	// Replace any null characters in the string. We would have expected strings.ToValidUTF8 to take
	// care of replacing any null characters, but it seems like this character is treated as valid a
	// UTF-8 character. See
	// https://stackoverflow.com/questions/6907297/can-utf-8-contain-zero-byte/6907327#6907327.

	// And it only appears that Postgres has a different idea of UTF-8 (only slightly). Without
	// using this function call, inserts for this string in Postgres return the following error:
	//
	// ERROR: invalid byte sequence for encoding "UTF8": 0x00 (SQLSTATE 22021)
	t := strings.ReplaceAll(s, "\x00", "")

	// Sanitize to a valid UTF-8 string and return it.
	return strings.ToValidUTF8(t, "")
}
