package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"golang.org/x/time/rate"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitserverRepoStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) GitserverRepoStore
	Upsert(ctx context.Context, repos ...*types.GitserverRepo) error
	IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error
	GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error)
	GetByName(ctx context.Context, name api.RepoName) (*types.GitserverRepo, error)
	GetByNames(ctx context.Context, names ...api.RepoName) ([]*types.GitserverRepo, error)
	SetCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus, shardID string) error
	SetLastError(ctx context.Context, name api.RepoName, error, shardID string) error
	SetLastFetched(ctx context.Context, name api.RepoName, data GitserverFetchData) error
	SetRepoSize(ctx context.Context, name api.RepoName, size int64, shardID string) error
	IterateWithNonemptyLastError(ctx context.Context, repoFn func(repo types.RepoGitserverStatus) error) error
	IteratePurgeableRepos(ctx context.Context, options IteratePurgableReposOptions, repoFn func(repo api.RepoName) error) error
	TotalErroredCloudDefaultRepos(ctx context.Context) (int, error)
	ListReposWithoutSize(ctx context.Context) (map[api.RepoName]api.RepoID, error)
	UpdateRepoSizes(ctx context.Context, shardID string, repos map[api.RepoID]int64) error
}

var _ GitserverRepoStore = (*gitserverRepoStore)(nil)

// gitserverRepoStore is responsible for data stored in the gitserver_repos table.
type gitserverRepoStore struct {
	*basestore.Store
}

// GitserverReposWith instantiates and returns a new gitserverRepoStore using
// the other store handle.
func GitserverReposWith(other basestore.ShareableStore) GitserverRepoStore {
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
		q := sqlf.Sprintf("(%s, %s, %s, %s, %s, %s, %s, now())",
			gr.RepoID,
			gr.CloneStatus,
			dbutil.NewNullString(gr.ShardID),
			dbutil.NewNullString(sanitizeToUTF8(gr.LastError)),
			gr.LastFetched,
			gr.LastChanged,
			gr.RepoSizeBytes,
		)

		values = append(values, q)
	}

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.Upsert
INSERT INTO
    gitserver_repos(repo_id, clone_status, shard_id, last_error, last_fetched, last_changed, repo_size_bytes, updated_at)
    VALUES %s
    ON CONFLICT (repo_id) DO UPDATE
    SET (clone_status, shard_id, last_error, last_fetched, last_changed, repo_size_bytes, updated_at) =
        (EXCLUDED.clone_status, EXCLUDED.shard_id, EXCLUDED.last_error, EXCLUDED.last_fetched, EXCLUDED.last_changed, EXCLUDED.repo_size_bytes, now())
`, sqlf.Join(values, ",")))

	return errors.Wrap(err, "upserting GitserverRepo")
}

// TotalErroredCloudDefaultRepos returns the total number of repos which have a non-empty last_error field. Note that this is only
// counting repos with an associated cloud_default external service.
func (s *gitserverRepoStore) TotalErroredCloudDefaultRepos(ctx context.Context) (int, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(totalErroredQuery))
	if err != nil {
		return 0, errors.Wrap(err, "fetching count of total errored repos")
	}
	var total int
	for rows.Next() {
		if err := rows.Scan(
			&total,
		); err != nil {
			return 0, errors.Wrap(err, "scanning row")
		}
	}
	return total, nil
}

const totalErroredQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.TotalErroredCloudDefaultRepos
SELECT
	count(*)
FROM repo
	INNER JOIN gitserver_repos gr ON repo.id = gr.repo_id
	INNER JOIN external_service_repos esr ON repo.id = esr.repo_id
	INNER JOIN external_services es on esr.external_service_id = es.id
WHERE gr.last_error != '' AND repo.deleted_at is NULL AND es.cloud_default IS True
`

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
	INNER JOIN gitserver_repos gr ON repo.id = gr.repo_id
	INNER JOIN external_service_repos esr ON repo.id = esr.repo_id
	INNER JOIN external_services es on esr.external_service_id = es.id
WHERE gr.last_error != '' AND repo.deleted_at is NULL AND es.cloud_default IS True
`

type IteratePurgableReposOptions struct {
	// DeletedBefore will filter the deleted repos to only those that were deleted
	// before the given time. The zero value will not apply filtering.
	DeletedBefore time.Time
	// Limit optionally limtits the repos iterated over. The zero value means no
	// limits are applied. Repos are ordered by their deleted at date, oldest first.
	Limit int
	// Limiter is an optional rate limiter that limits the rate at which we iterate
	// through the repos.
	Limiter *rate.Limiter
}

// IteratePurgeableRepos iterates over all purgeable repos. These are repos that
// are cloned on disk but have been deleted or blocked.
func (s *gitserverRepoStore) IteratePurgeableRepos(ctx context.Context, options IteratePurgableReposOptions, repoFn func(repo api.RepoName) error) error {
	deletedAtClause := sqlf.Sprintf("deleted_at IS NOT NULL")
	if !options.DeletedBefore.IsZero() {
		deletedAtClause = sqlf.Sprintf("(deleted_at IS NOT NULL AND deleted_at < %s)", options.DeletedBefore)
	}
	query := purgableReposQuery
	if options.Limit > 0 {
		query = query + fmt.Sprintf(" LIMIT %d", options.Limit)
	}
	rows, err := s.Query(ctx, sqlf.Sprintf(query, deletedAtClause))
	if err != nil {
		return errors.Wrap(err, "fetching repos with nonempty last_error")
	}
	defer rows.Close()

	for rows.Next() {
		var name api.RepoName
		if err := rows.Scan(
			&name,
		); err != nil {
			return errors.Wrap(err, "scanning row")
		}
		err := repoFn(name)
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

const purgableReposQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IteratePurgeableRepos
SELECT
	repo.name
FROM repo
	JOIN gitserver_repos gr ON repo.id = gr.repo_id
WHERE (%s OR repo.blocked IS NOT NULL)
AND gr.clone_status = 'cloned'
ORDER BY deleted_at asc
`

type IterateRepoGitserverStatusOptions struct {
	// If set, will only iterate over repos that have not been assigned to a shard
	OnlyWithoutShard bool
	// If true, also include deleted repos. Note that their repo name will start with
	// 'DELETED-'
	IncludeDeleted bool
}

// IterateRepoGitserverStatus iterates over the status of all repos by joining
// our repo and gitserver_repos table. It is impossible for us not to have a
// corresponding row in gitserver_repos because of the trigger on repos table.
// repoFn will be called once for each row. If it returns an error we'll abort iteration.
func (s *gitserverRepoStore) IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error {
	if repoFn == nil {
		return errors.New("nil repoFn")
	}

	deletedClause := sqlf.Sprintf("repo.deleted_at is null")
	if options.IncludeDeleted {
		deletedClause = sqlf.Sprintf("TRUE")
	}

	var q *sqlf.Query
	if options.OnlyWithoutShard {
		q = sqlf.Sprintf(iterateRepoGitserverStatusWithoutShardQuery, deletedClause, deletedClause)
	} else {
		q = sqlf.Sprintf(iterateRepoGitserverQuery, deletedClause)
	}

	rows, err := s.Query(ctx, q)
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
			&dbutil.NullInt64{N: &gr.RepoSizeBytes},
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
	gr.repo_size_bytes,
	gr.updated_at
FROM repo
LEFT JOIN gitserver_repos gr ON gr.repo_id = repo.id
WHERE %s
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
		NULL AS repo_size_bytes,
		NULL AS updated_at
	FROM repo
	WHERE %s AND NOT EXISTS (SELECT 1 FROM gitserver_repos gr WHERE gr.repo_id = repo.id)
) UNION ALL (
	SELECT
		repo.id,
		repo.name,
		gr.clone_status,
		gr.shard_id,
		gr.last_error,
		gr.last_fetched,
		gr.last_changed,
		gr.repo_size_bytes,
		gr.updated_at
	FROM repo
	JOIN gitserver_repos gr ON gr.repo_id = repo.id
	WHERE %s AND gr.shard_id = ''
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
	   repo_size_bytes,
       updated_at
FROM gitserver_repos
WHERE repo_id = %s
`

	return scanSingleGitserverRepo(s.QueryRow(ctx, sqlf.Sprintf(q, id)))
}

func (s *gitserverRepoStore) GetByName(ctx context.Context, name api.RepoName) (*types.GitserverRepo, error) {
	q := `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByName
SELECT
       g.repo_id,
       g.clone_status,
       g.shard_id,
       g.last_error,
       g.last_fetched,
       g.last_changed,
	   g.repo_size_bytes,
       g.updated_at
FROM gitserver_repos g
JOIN repo r on r.id = g.repo_id
WHERE r.name = %s
`

	return scanSingleGitserverRepo(s.QueryRow(ctx, sqlf.Sprintf(q, name)))
}

const getByNamesQueryTemplate = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByName
SELECT
       g.repo_id,
       g.clone_status,
       g.shard_id,
       g.last_error,
       g.last_fetched,
       g.last_changed,
	   g.repo_size_bytes,
       g.updated_at
FROM gitserver_repos g
JOIN repo r on r.id = g.repo_id
WHERE r.name IN (%s)
`

func (s *gitserverRepoStore) GetByNames(ctx context.Context, names ...api.RepoName) ([]*types.GitserverRepo, error) {
	return s.getByNames(ctx, batch.MaxNumPostgresParameters, names...)
}

func (s *gitserverRepoStore) getByNames(ctx context.Context, maxNumPostgresParameters int, names ...api.RepoName) ([]*types.GitserverRepo, error) {
	remainingNames := len(names)
	nameQueries := make([]*sqlf.Query, 0)
	batchSize := 0
	repos := make([]*types.GitserverRepo, 0, remainingNames)

	// iterating len(names) + 1 times because last iteration is needed for the last batch
	for i := 0; i <= len(names); i++ {
		if remainingNames == 0 || batchSize == maxNumPostgresParameters {
			// executing the DB query
			res, err := s.sendBatchQuery(ctx, batchSize, nameQueries)
			if err != nil {
				return nil, err
			}
			repos = append(repos, res...)

			if remainingNames == 0 {
				// last batch: break out of the loop
				break
			}

			// intermediate batch: reset variables required for a new batch
			batchSize = 0
			nameQueries = nil
		}
		nameQueries = append(nameQueries, sqlf.Sprintf("%s", names[i]))
		batchSize++
		remainingNames--
	}

	return repos, nil
}

func (s *gitserverRepoStore) sendBatchQuery(ctx context.Context, batchSize int, nameQueries []*sqlf.Query) ([]*types.GitserverRepo, error) {
	repos := make([]*types.GitserverRepo, 0, batchSize)
	rows, err := s.Query(ctx, sqlf.Sprintf(getByNamesQueryTemplate, sqlf.Join(nameQueries, ",")))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		repo, err := scanSingleGitserverRepo(rows)
		if err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, err
}

// ScannerWithError captures Scan and Err methods of sql.Rows and sql.Row.
type ScannerWithError interface {
	Scan(dst ...any) error
	Err() error
}

func scanSingleGitserverRepo(scanner ScannerWithError) (*types.GitserverRepo, error) {
	if scanner.Err() != nil {
		return nil, errors.Wrap(scanner.Err(), "getting GitserverRepo")
	}
	var gr types.GitserverRepo
	var cloneStatus string
	err := scanner.Scan(
		&gr.RepoID,
		&cloneStatus,
		&gr.ShardID,
		&dbutil.NullString{S: &gr.LastError},
		&dbutil.NullTime{Time: &gr.LastFetched},
		&dbutil.NullTime{Time: &gr.LastChanged},
		&dbutil.NullInt64{N: &gr.RepoSizeBytes},
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

// SetRepoSize will attempt to update ONLY the repo size of a GitServerRepo. If
// a matching row does not yet exist a new one will be created.
// If the size value hasn't changed, the row will not be updated.
func (s *gitserverRepoStore) SetRepoSize(ctx context.Context, name api.RepoName, size int64, shardID string) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
	-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetRepoSize
	INSERT INTO gitserver_repos(repo_id, repo_size_bytes, shard_id, updated_at)
	SELECT id, %s, %s, now()
	FROM repo
	WHERE name = %s
	ON CONFLICT (repo_id) DO UPDATE
	       SET (repo_size_bytes, updated_at) =
	                       (EXCLUDED.repo_size_bytes, now())
	WHERE gitserver_repos.repo_size_bytes IS DISTINCT FROM EXCLUDED.repo_size_bytes
	`, size, shardID, name))

	return errors.Wrap(err, "setting repo size")
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

// ListReposWithoutSize returns a map of repo name to repo ID for repos which do not have a repo_size_bytes
func (s *gitserverRepoStore) ListReposWithoutSize(ctx context.Context) (map[api.RepoName]api.RepoID, error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(listReposWithoutSizeQuery))
	if err != nil {
		return nil, errors.Wrap(err, "fetching repos without size")
	}
	defer rows.Close()
	repos := make(map[api.RepoName]api.RepoID, 0)
	for rows.Next() {
		var name string
		var ID int32
		if err := rows.Scan(&name, &ID); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		repos[api.RepoName(name)] = api.RepoID(ID)
	}
	return repos, nil
}

const listReposWithoutSizeQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.ListReposWithoutSize
SELECT
	repo.name,
    repo.id
FROM repo
JOIN gitserver_repos gr ON gr.repo_id = repo.id
WHERE gr.repo_size_bytes IS NULL
`

// UpdateRepoSizes sets repo sizes according to input map. Key is repoID, value is repo_size_bytes.
func (s *gitserverRepoStore) UpdateRepoSizes(ctx context.Context, shardID string, repos map[api.RepoID]int64) (err error) {

	inserter := func(inserter *batch.Inserter) error {
		for repo, size := range repos {
			if err := inserter.Insert(ctx, repo, shardID, size, "now()"); err != nil {
				return err
			}
		}
		return nil
	}

	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return err
	}
	defer func() { err = tx.Done(err) }()

	if err := batch.WithInserterWithReturn(
		ctx,
		tx.Handle(),
		"gitserver_repos",
		batch.MaxNumPostgresParameters,
		[]string{"repo_id", "shard_id", "repo_size_bytes", "updated_at"},
		"ON CONFLICT (repo_id) DO UPDATE SET (repo_size_bytes, shard_id, updated_at) = (EXCLUDED.repo_size_bytes, gitserver_repos.shard_id, now())",
		nil,
		nil,
		inserter,
	); err != nil {
		return err
	}
	return nil
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
