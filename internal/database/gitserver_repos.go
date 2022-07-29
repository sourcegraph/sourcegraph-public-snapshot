package database

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/batch"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/ratelimit"
	"github.com/sourcegraph/sourcegraph/internal/types"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

type GitserverRepoStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) GitserverRepoStore

	// Update updates the given rows with the GitServer status of a repo.
	Update(ctx context.Context, repos ...*types.GitserverRepo) error
	// IterateRepoGitserverStatus iterates over the status of all repos by joining
	// our repo and gitserver_repos table. It is impossible for us not to have a
	// corresponding row in gitserver_repos because of the trigger on repos table.
	// repoFn will be called once for each row. If it returns an error we'll abort iteration.
	IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) error
	GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error)
	GetByName(ctx context.Context, name api.RepoName) (*types.GitserverRepo, error)
	GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error)
	// SetCloneStatus will attempt to update ONLY the clone status of a
	// GitServerRepo. If a matching row does not yet exist a new one will be created.
	// If the status value hasn't changed, the row will not be updated.
	SetCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus, shardID string) error
	// SetLastError will attempt to update ONLY the last error of a GitServerRepo. If
	// a matching row does not yet exist a new one will be created.
	// If the error value hasn't changed, the row will not be updated.
	SetLastError(ctx context.Context, name api.RepoName, error, shardID string) error
	// SetLastFetched will attempt to update ONLY the last fetched data (last_fetched, last_changed, shard_id) of a GitServerRepo and ensures it is marked as cloned.
	SetLastFetched(ctx context.Context, name api.RepoName, data GitserverFetchData) error
	// SetRepoSize will attempt to update ONLY the repo size of a GitServerRepo. If
	// a matching row does not yet exist a new one will be created.
	// If the size value hasn't changed, the row will not be updated.
	SetRepoSize(ctx context.Context, name api.RepoName, size int64, shardID string) error
	// IterateWithNonemptyLastError iterates over repos w/ non-empty last_error field and calls the repoFn for these repos.
	// note that this currently filters out any repos which do not have an associated external service where cloud_default = true.
	IterateWithNonemptyLastError(ctx context.Context, repoFn func(repo api.RepoName) error) error
	// IteratePurgeableRepos iterates over all purgeable repos. These are repos that
	// are cloned on disk but have been deleted or blocked.
	IteratePurgeableRepos(ctx context.Context, options IteratePurgableReposOptions, repoFn func(repo api.RepoName) error) error
	// TotalErroredCloudDefaultRepos returns the total number of repos which have a non-empty last_error field. Note that this is only
	// counting repos with an associated cloud_default external service.
	TotalErroredCloudDefaultRepos(ctx context.Context) (int, error)
	// ListReposWithoutSize returns a map of repo name to repo ID for repos which do not have a repo_size_bytes.
	ListReposWithoutSize(ctx context.Context) (map[api.RepoName]api.RepoID, error)
	// UpdateRepoSizes sets repo sizes according to input map. Key is repoID, value is repo_size_bytes.
	UpdateRepoSizes(ctx context.Context, shardID string, repos map[api.RepoID]int64) (int, error)
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

func (s *gitserverRepoStore) Update(ctx context.Context, repos ...*types.GitserverRepo) error {
	values := make([]*sqlf.Query, 0, len(repos))
	for _, gr := range repos {
		values = append(values, sqlf.Sprintf("(%s::integer, %s::text, %s::text, %s::text, %s::timestamp with time zone, %s::timestamp with time zone, %s::bigint, NOW())",
			gr.RepoID,
			gr.CloneStatus,
			gr.ShardID,
			dbutil.NewNullString(sanitizeToUTF8(gr.LastError)),
			gr.LastFetched,
			gr.LastChanged,
			&dbutil.NullInt64{N: &gr.RepoSizeBytes},
		))
	}

	err := s.Exec(ctx, sqlf.Sprintf(updateGitserverReposQueryFmtstr, sqlf.Join(values, ",")))

	return errors.Wrap(err, "updating GitserverRepo")
}

const updateGitserverReposQueryFmtstr = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.Update
UPDATE gitserver_repos AS gr
SET
	clone_status = tmp.clone_status,
	shard_id = tmp.shard_id,
	last_error = tmp.last_error,
	last_fetched = tmp.last_fetched,
	last_changed = tmp.last_changed,
	repo_size_bytes = tmp.repo_size_bytes,
	updated_at = NOW()
FROM (VALUES
	-- (<repo_id>, <clone_status>, <shard_id>, <last_error>, <last_fetched>, <last_changed>, <repo_size_bytes>),
		%s
	) AS tmp(repo_id, clone_status, shard_id, last_error, last_fetched, last_changed, repo_size_bytes)
	WHERE
		tmp.repo_id = gr.repo_id
`

func (s *gitserverRepoStore) TotalErroredCloudDefaultRepos(ctx context.Context) (int, error) {
	count, _, err := basestore.ScanFirstInt(s.Query(ctx, sqlf.Sprintf(totalErroredCloudDefaultReposQuery)))
	return count, err
}

const totalErroredCloudDefaultReposQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.TotalErroredCloudDefaultRepos
SELECT
	COUNT(*)
FROM gitserver_repos gr
JOIN repo r ON r.id = gr.repo_id
JOIN external_service_repos esr ON gr.repo_id = esr.repo_id
JOIN external_services es on esr.external_service_id = es.id
WHERE
	gr.last_error != ''
	AND r.deleted_at IS NULL
	AND es.cloud_default IS TRUE
`

func (s *gitserverRepoStore) IterateWithNonemptyLastError(ctx context.Context, repoFn func(repo api.RepoName) error) (err error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(nonemptyLastErrorQuery))
	if err != nil {
		return errors.Wrap(err, "fetching repos with nonempty last_error")
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	for rows.Next() {
		var name api.RepoName
		if err := rows.Scan(&name); err != nil {
			return errors.Wrap(err, "scanning row")
		}
		err := repoFn(name)
		if err != nil {
			// Abort
			return errors.Wrap(err, "calling repoFn")
		}
	}

	return nil
}

const nonemptyLastErrorQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IterateWithNonemptyLastError
SELECT
	repo.name
FROM repo
JOIN gitserver_repos gr ON repo.id = gr.repo_id
JOIN external_service_repos esr ON repo.id = esr.repo_id
JOIN external_services es on esr.external_service_id = es.id
WHERE
	gr.last_error != ''
	AND repo.deleted_at IS NULL
	AND es.cloud_default IS TRUE
`

type IteratePurgableReposOptions struct {
	// DeletedBefore will filter the deleted repos to only those that were deleted
	// before the given time. The zero value will not apply filtering.
	DeletedBefore time.Time
	// Limit optionally limits the repos iterated over. The zero value means no
	// limits are applied. Repos are ordered by their deleted at date, oldest first.
	Limit int
	// Limiter is an optional rate limiter that limits the rate at which we iterate
	// through the repos.
	Limiter *ratelimit.InstrumentedLimiter
}

func (s *gitserverRepoStore) IteratePurgeableRepos(ctx context.Context, options IteratePurgableReposOptions, repoFn func(repo api.RepoName) error) (err error) {
	deletedAtClause := sqlf.Sprintf("deleted_at IS NOT NULL")
	if !options.DeletedBefore.IsZero() {
		deletedAtClause = sqlf.Sprintf("(deleted_at IS NOT NULL AND deleted_at < %s)", options.DeletedBefore)
	}
	query := purgableReposQuery
	if options.Limit > 0 {
		query = query + fmt.Sprintf(" LIMIT %d", options.Limit)
	}
	rows, err := s.Query(ctx, sqlf.Sprintf(query, deletedAtClause, types.CloneStatusCloned))
	if err != nil {
		return errors.Wrap(err, "fetching repos with nonempty last_error")
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	for rows.Next() {
		var name api.RepoName
		if err := rows.Scan(&name); err != nil {
			return errors.Wrap(err, "scanning row")
		}
		err := repoFn(name)
		if err != nil {
			// Abort
			return errors.Wrap(err, "calling repoFn")
		}
	}

	return nil
}

const purgableReposQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IteratePurgeableRepos
SELECT
	repo.name
FROM repo
JOIN gitserver_repos gr ON repo.id = gr.repo_id
WHERE
	(%s OR repo.blocked IS NOT NULL)
	AND gr.clone_status = %s
ORDER BY deleted_at ASC
`

type IterateRepoGitserverStatusOptions struct {
	// If set, will only iterate over repos that have not been assigned to a shard
	OnlyWithoutShard bool
	// If true, also include deleted repos. Note that their repo name will start with
	// 'DELETED-'
	IncludeDeleted bool
}

func (s *gitserverRepoStore) IterateRepoGitserverStatus(ctx context.Context, options IterateRepoGitserverStatusOptions, repoFn func(repo types.RepoGitserverStatus) error) (err error) {
	if repoFn == nil {
		return errors.New("nil repoFn")
	}

	preds := []*sqlf.Query{}

	if !options.IncludeDeleted {
		preds = append(preds, sqlf.Sprintf("repo.deleted_at IS NULL"))
	}

	if options.OnlyWithoutShard {
		preds = append(preds, sqlf.Sprintf("gr.shard_id = ''"))
	}

	if len(preds) == 0 {
		preds = append(preds, sqlf.Sprintf("TRUE"))
	}

	q := sqlf.Sprintf(iterateRepoGitserverQuery, sqlf.Join(preds, "AND"))

	rows, err := s.Query(ctx, q)
	if err != nil {
		return errors.Wrap(err, "fetching gitserver status")
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

	for rows.Next() {
		gr, name, err := scanGitserverRepo(rows)
		if err != nil {
			return errors.Wrap(err, "scanning row")
		}

		rgs := types.RepoGitserverStatus{
			ID:            gr.RepoID,
			Name:          name,
			GitserverRepo: gr,
		}

		if err := repoFn(rgs); err != nil {
			// Abort
			return errors.Wrap(err, "calling repoFn")
		}
	}

	return nil
}

const iterateRepoGitserverQuery = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.IterateRepoGitserverStatus
SELECT
	gr.repo_id,
	repo.name,
	gr.clone_status,
	gr.shard_id,
	gr.last_error,
	gr.last_fetched,
	gr.last_changed,
	gr.repo_size_bytes,
	gr.updated_at
FROM gitserver_repos gr
JOIN repo ON gr.repo_id = repo.id
WHERE %s
`

func (s *gitserverRepoStore) GetByID(ctx context.Context, id api.RepoID) (*types.GitserverRepo, error) {
	repo, _, err := scanGitserverRepo(s.QueryRow(ctx, sqlf.Sprintf(getGitserverRepoByIDQueryFmtstr, id)))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errGitserverRepoNotFound{}
		}
	}
	return repo, nil
}

const getGitserverRepoByIDQueryFmtstr = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByID
SELECT
	repo_id,
	-- We don't need this here, but the scanner needs it.
	'' as name,
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

func (s *gitserverRepoStore) GetByName(ctx context.Context, name api.RepoName) (*types.GitserverRepo, error) {
	repo, _, err := scanGitserverRepo(s.QueryRow(ctx, sqlf.Sprintf(getGitserverRepoByNameQueryFmtstr, name)))
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, &errGitserverRepoNotFound{}
		}
	}
	return repo, nil
}

const getGitserverRepoByNameQueryFmtstr = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByName
SELECT
	gr.repo_id,
	-- We don't need this here, but the scanner needs it.
	'' as name,
	gr.clone_status,
	gr.shard_id,
	gr.last_error,
	gr.last_fetched,
	gr.last_changed,
	gr.repo_size_bytes,
	gr.updated_at
FROM gitserver_repos gr
JOIN repo r ON r.id = gr.repo_id
WHERE r.name = %s
`

type errGitserverRepoNotFound struct{}

func (err *errGitserverRepoNotFound) Error() string { return "gitserver repo not found" }
func (errGitserverRepoNotFound) NotFound() bool     { return true }

const getByNamesQueryTemplate = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.GetByNames
SELECT
	gr.repo_id,
	r.name,
	gr.clone_status,
	gr.shard_id,
	gr.last_error,
	gr.last_fetched,
	gr.last_changed,
	gr.repo_size_bytes,
	gr.updated_at
FROM gitserver_repos gr
JOIN repo r on r.id = gr.repo_id
WHERE r.name = ANY (%s)
`

func (s *gitserverRepoStore) GetByNames(ctx context.Context, names ...api.RepoName) (map[api.RepoName]*types.GitserverRepo, error) {
	repos := make(map[api.RepoName]*types.GitserverRepo, len(names))

	rows, err := s.Query(ctx, sqlf.Sprintf(getByNamesQueryTemplate, pq.Array(names)))
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	for rows.Next() {
		repo, repoName, err := scanGitserverRepo(rows)
		if err != nil {
			return nil, err
		}
		repos[repoName] = repo
	}

	return repos, nil
}

func scanGitserverRepo(scanner dbutil.Scanner) (*types.GitserverRepo, api.RepoName, error) {
	var gr types.GitserverRepo
	var cloneStatus string
	var repoName api.RepoName
	err := scanner.Scan(
		&gr.RepoID,
		&repoName,
		&cloneStatus,
		&gr.ShardID,
		&dbutil.NullString{S: &gr.LastError},
		&gr.LastFetched,
		&gr.LastChanged,
		&dbutil.NullInt64{N: &gr.RepoSizeBytes},
		&gr.UpdatedAt,
	)
	if err != nil {
		return nil, "", errors.Wrap(err, "scanning GitserverRepo")
	}
	gr.CloneStatus = types.ParseCloneStatus(cloneStatus)

	return &gr, repoName, nil
}

func (s *gitserverRepoStore) SetCloneStatus(ctx context.Context, name api.RepoName, status types.CloneStatus, shardID string) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetCloneStatus
UPDATE gitserver_repos
SET
	clone_status = %s,
	shard_id = %s,
	updated_at = NOW()
WHERE
	repo_id = (SELECT id FROM repo WHERE name = %s)
	AND
	clone_status IS DISTINCT FROM %s
`, status, shardID, name, status))
	if err != nil {
		return errors.Wrap(err, "setting clone status")
	}

	return nil
}

func (s *gitserverRepoStore) SetLastError(ctx context.Context, name api.RepoName, error, shardID string) error {
	ns := dbutil.NewNullString(sanitizeToUTF8(error))

	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetLastError
UPDATE gitserver_repos
SET
	last_error = %s,
	shard_id = %s,
	updated_at = NOW()
WHERE
	repo_id = (SELECT id FROM repo WHERE name = %s)
	AND
	last_error IS DISTINCT FROM %s
`, ns, shardID, name, ns))
	if err != nil {
		return errors.Wrap(err, "setting last error")
	}

	return nil
}

func (s *gitserverRepoStore) SetRepoSize(ctx context.Context, name api.RepoName, size int64, shardID string) error {
	err := s.Exec(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetRepoSize
UPDATE gitserver_repos
SET
	repo_size_bytes = %s,
	shard_id = %s,
	clone_status = %s,
	updated_at = NOW()
WHERE
	repo_id = (SELECT id FROM repo WHERE name = %s)
	AND
	repo_size_bytes IS DISTINCT FROM %s
	`, size, shardID, types.CloneStatusCloned, name, size))
	if err != nil {
		return errors.Wrap(err, "setting repo size")
	}

	return nil
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

func (s *gitserverRepoStore) SetLastFetched(ctx context.Context, name api.RepoName, data GitserverFetchData) error {
	res, err := s.ExecResult(ctx, sqlf.Sprintf(`
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.SetLastFetched
UPDATE gitserver_repos
SET
	last_fetched = %s,
	last_changed = %s,
	shard_id = %s,
	clone_status = %s,
	updated_at = NOW()
WHERE repo_id = (SELECT id FROM repo WHERE name = %s)
`, data.LastFetched, data.LastChanged, data.ShardID, types.CloneStatusCloned, name))
	if err != nil {
		return errors.Wrap(err, "setting last fetched")
	}

	if nrows, err := res.RowsAffected(); err != nil {
		return errors.Wrap(err, "getting rows affected")
	} else if nrows != 1 {
		return errors.New("repo not found")
	}

	return nil
}

func (s *gitserverRepoStore) ListReposWithoutSize(ctx context.Context) (_ map[api.RepoName]api.RepoID, err error) {
	rows, err := s.Query(ctx, sqlf.Sprintf(listReposWithoutSizeQuery))
	if err != nil {
		return nil, errors.Wrap(err, "fetching repos without size")
	}
	defer func() {
		err = basestore.CloseRows(rows, err)
	}()

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

func (s *gitserverRepoStore) UpdateRepoSizes(ctx context.Context, shardID string, repos map[api.RepoID]int64) (updated int, err error) {
	// NOTE: We have two args per row, so rows*2 should be less than maximum
	// Postgres allows.
	const batchSize = batch.MaxNumPostgresParameters / 2
	return s.updateRepoSizesWithBatchSize(ctx, shardID, repos, batchSize)
}

func (s *gitserverRepoStore) updateRepoSizesWithBatchSize(ctx context.Context, shardID string, repos map[api.RepoID]int64, batchSize int) (updated int, err error) {
	tx, err := s.Store.Transact(ctx)
	if err != nil {
		return 0, err
	}
	defer func() { err = tx.Done(err) }()

	batch := make([]*sqlf.Query, batchSize)

	left := len(repos)
	currentCount := 0
	updatedRows := 0
	for repo, size := range repos {
		batch[currentCount] = sqlf.Sprintf("(%s::integer, %s::bigint)", repo, size)

		currentCount += 1

		if currentCount == batchSize || currentCount == left {
			// IMPORTANT: we only take the elements of batch up to currentCount
			q := sqlf.Sprintf(updateRepoSizesQueryFmtstr, sqlf.Join(batch[:currentCount], ","))
			res, err := tx.ExecResult(ctx, q)
			if err != nil {
				return 0, err
			}

			rowsAffected, err := res.RowsAffected()
			if err != nil {
				return 0, err
			}
			updatedRows += int(rowsAffected)

			left -= currentCount
			currentCount = 0
		}
	}

	return updatedRows, nil
}

const updateRepoSizesQueryFmtstr = `
-- source: internal/database/gitserver_repos.go:gitserverRepoStore.UpdateRepoSizes
UPDATE gitserver_repos AS gr
SET
    repo_size_bytes = tmp.repo_size_bytes
FROM (VALUES
-- (<repo_id>, <repo_size_bytes>),
    %s
) AS tmp(repo_id, repo_size_bytes)
WHERE
	tmp.repo_id = gr.repo_id
AND
	tmp.repo_size_bytes IS DISTINCT FROM gr.repo_size_bytes
`

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
