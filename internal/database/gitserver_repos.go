package database

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
)

type GitserverRepoStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) GitserverRepoStore
	Upsert(ctx context.Context, repos ...*types.GitserverRepo) error
	IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error
	GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error)
	SetCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus, shardID string) error
	SetLastError(ctx context.Context, name api.RepoName, error, shardID string) error
	SetLastFetched(ctx context.Context, name api.RepoName, data GitserverFetchData) error
	IterateWithNonemptyLastError(ctx context.Context, repoFn func(repo types.RepoGitserverStatus) error) error
}

var _ GitserverRepoStore = (*gitserverRepoStore)(nil)

// gitserverRepoStore is responsible for data stored in the gitserver_repos table.
type gitserverRepoStore struct {
	*basestore.Store
}

// GitserverRepos instantiates and returns a new gitserverRepoStore.
func GitserverRepos(db dbutil.DB) GitserverRepoStore {
	return &gitserverRepoStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// NewGitserverReposWith instantiates and returns a new gitserverRepoStore using
// the other store handle.
func NewGitserverReposWith(other basestore.ShareableStore) GitserverRepoStore {
	return &gitserverRepoStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *gitserverRepoStore) With(other basestore.ShareableStore) GitserverRepoStore {
	return &gitserverRepoStore{Store: s.Store.With(other)}
}

func (s *gitserverRepoStore) Transact(ctx context.Context) (GitserverRepoStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &gitserverRepoStore{Store: txBase}, err
}

// Upsert adds a row representing the GitServer status of a repo
func (s *gitserverRepoStore) Upsert(ctx context.Context, repos ...*types.GitserverRepo) error {
	values := make([]*sqlf.Query, 0, len(repos))
	for _, gr := range repos {
		q := sqlf.Sprintf("(%s, %s, %s, %s, %s, %s, now())",
			gr.RepoID,
			gr.CloneStatus,
			dbutil.NewNullString(gr.ShardID),
			dbutil.NewNullString(sanitizeToUTF8(gr.LastError)),
			gr.LastFetched,
			gr.LastChanged,
		)

		values = append(values, q)
	}

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.Upsert
INSERT INTO
    gitserver_repos(repo_id, clone_status, shard_id, last_error, last_fetched, last_changed, updated_at)
    VALUES %s
    ON CONFLICT (repo_id) DO UPDATE
    SET (clone_status, shard_id, last_error, last_fetched, last_changed, updated_at) =
        (EXCLUDED.clone_status, EXCLUDED.shard_id, EXCLUDED.last_error, EXCLUDED.last_fetched, EXCLUDED.last_changed, now())
`, sqlf.Join(values, ",")))

	return errors.Wrap(err, "creating GitserverRepo")
}

// IterateWithNonemptyLastError iterates over repos w/ non-empty last_error field and calls the repoFn for these repos.
// note that this currently filters out any repos which do not have an associated external service where cloud_default = true.
func (s *gitserverRepoStore) IterateWithNonemptyLastError(ctx context.Context, repoFn func(repo types.RepoGitserverStatus) error) error {
	rows, err := s.Query(ctx, sqlf.Sprintf(nonemptyLastErrorQuery))
	if err != nil {
		return errors.Wrap(err, "fetching repos with nonempty last_error")
	}
	defer rows.Close()

	for rows.Next() {
		var gr types.GitserverRepo
		var rgs types.RepoGitserverStatus
		if err := rows.Scan(
			&rgs.Name,
			&dbutil.NullString{S: &gr.LastError},
		); err != nil {
			return errors.Wrap(err, "scanning row")
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

const nonemptyLastErrorQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IterateWithNonemptyLastError
SELECT
	repo.name,
	gr.last_error
FROM repo
	LEFT JOIN gitserver_repos gr ON repo.id = gr.repo_id
	INNER JOIN external_service_repos esr ON repo.id = esr.repo_id
	INNER JOIN external_services es on esr.external_service_id = es.id
WHERE gr.last_error != '' AND repo.deleted_at is NULL AND es.cloud_default IS True
`

type IterateRepoGitserverStatusOptions struct {
	// If set, will only iterate over repos that have not been assigned to a shard
	OnlyWithoutShard bool
}

// IterateRepoGitserverStatus iterates over the status of all repos by joining
// our repo and gitserver_repos table. It is possible for us not to have a
// corresponding row in gitserver_repos yet. repoFn will be called once for each
// row. If it returns an error we'll abort iteration.
func (s *gitserverRepoStore) IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error {
	if repoFn == nil {
		return errors.New("nil repoFn")
	}

	var q string
	if options.OnlyWithoutShard {
		q = iterateRepoGitserverStatusWithoutShardQuery
	} else {
		q = iterateRepoGitserverQuery
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
			&dbutil.NullString{S: &gr.LastError},
			&dbutil.NullTime{Time: &gr.LastFetched},
			&dbutil.NullTime{Time: &gr.LastChanged},
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

const iterateRepoGitserverQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IterateRepoGitserverStatus
SELECT
	repo.id,
	repo.name,
	gr.clone_status,
	gr.shard_id,
	gr.last_error,
	gr.last_fetched,
	gr.last_changed,
	gr.updated_at
FROM repo
LEFT JOIN gitserver_repos gr ON gr.repo_id = repo.id
WHERE repo.deleted_at IS NULL
`

const iterateRepoGitserverStatusWithoutShardQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IterateRepoGitserverStatus
(
	SELECT
		repo.id,
		repo.name,
		NULL AS clone_status,
		NULL AS shard_id,
		NULL AS last_error,
		NULL AS last_fetched,
		NULL AS last_changed,
		NULL AS updated_at
	FROM repo
	WHERE repo.deleted_at IS NULL AND NOT EXISTS (SELECT 1 FROM gitserver_repos gr WHERE gr.repo_id = repo.id)
) UNION ALL (
	SELECT
		repo.id,
		repo.name,
		gr.clone_status,
		gr.shard_id,
		gr.last_error,
		gr.last_fetched,
		gr.last_changed,
		gr.updated_at
	FROM repo
	JOIN gitserver_repos gr ON gr.repo_id = repo.id
	WHERE repo.deleted_at IS NULL AND gr.shard_id = ''
)
`

func (s *gitserverRepoStore) GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error) {
	q := `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByID
SELECT
       repo_id,
       clone_status,
       shard_id,
       last_error,
       last_fetched,
       last_changed,
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
		&dbutil.NullString{S: &gr.LastError},
		&dbutil.NullTime{Time: &gr.LastFetched},
		&dbutil.NullTime{Time: &gr.LastChanged},
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
func (s *gitserverRepoStore) SetCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus, shardID string) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetCloneStatus
INSERT INTO gitserver_repos(repo_id, clone_status, shard_id, updated_at)
SELECT id, %s, %s, now()
FROM repo
WHERE name = %s
ON CONFLICT (repo_id) DO UPDATE
SET (clone_status, shard_id, updated_at) =
    (EXCLUDED.clone_status, EXCLUDED.shard_id, now())
    WHERE gitserver_repos.clone_status IS DISTINCT FROM EXCLUDED.clone_status
`, status, shardID, name))

	return errors.Wrap(err, "setting clone status")
}

// SetLastError will attempt to update ONLY the last error of a GitServerRepo. If
// a matching row does not yet exist a new one will be created.
// If the error value hasn't changed, the row will not be updated.
func (s *gitserverRepoStore) SetLastError(ctx context.Context, name api.RepoName, error, shardID string) error {
	ns := dbutil.NewNullString(sanitizeToUTF8(error))

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetLastError
INSERT INTO gitserver_repos(repo_id, last_error, shard_id, updated_at)
SELECT id, %s, %s, now()
FROM repo
WHERE name = %s
ON CONFLICT (repo_id) DO UPDATE
    SET (last_error, shard_id, updated_at) =
            (EXCLUDED.last_error, EXCLUDED.shard_id, now())
WHERE gitserver_repos.last_error IS DISTINCT FROM EXCLUDED.last_error
`, ns, shardID, name))

	return errors.Wrap(err, "setting last error")
}

// GitserverFetchData is the metadata associated with a fetch operation on
// gitserver.
type GitserverFetchData struct {
	// LastFetched was the time the fetch operation completed (gitserver_repos.last_fetched).
	LastFetched time.Time
	// LastChanged was the last time a fetch changed the contents of the repo (gitserver_repos.last_changed).
	LastChanged time.Time
	// ShardID is the name of the gitserver the fetch ran on (gitserver.shard_id).
	ShardID string
}

// SetLastFetched will attempt to update ONLY the last fetched data of a GitServerRepo.
// a matching row does not yet exist a new one will be created.
func (s *gitserverRepoStore) SetLastFetched(ctx context.Context, name api.RepoName, data GitserverFetchData) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetLastFetched
INSERT INTO gitserver_repos(repo_id, last_fetched, last_changed, shard_id, updated_at)
SELECT id, %s, %s, %s, now()
FROM repo WHERE name = %s
ON CONFLICT (repo_id) DO UPDATE
SET (last_fetched, last_changed, shard_id, updated_at) =
    (EXCLUDED.last_fetched, EXCLUDED.last_changed, EXCLUDED.shard_id, now())
`, data.LastFetched, data.LastChanged, data.ShardID, name))

	return errors.Wrap(err, "setting last fetched")
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
