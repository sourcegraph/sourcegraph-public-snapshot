package database

import (
	"context"
	"database/sql"

	"github.com/cockroachdb/errors"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
)

// SubRepoPermsVersion is defines the version we are using to encode our include
// and exclude patterns.
const SubRepoPermsVersion = 1

var SubRepoSupportedCodeHostKinds = []string{extsvc.KindPerforce}
var supportedKindsQuery = make([]*sqlf.Query, len(SubRepoSupportedCodeHostKinds))

func init() {
	// Build this up at startup so we don't need to rebuild it every time
	// RepoSupported is called
	for i, kind := range SubRepoSupportedCodeHostKinds {
		supportedKindsQuery[i] = sqlf.Sprintf("%s", kind)
	}
}

type SubRepoPermsStore interface {
	With(other basestore.ShareableStore) SubRepoPermsStore
	Transact(ctx context.Context) (SubRepoPermsStore, error)
	Done(err error) error
	Upsert(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissions) error
	UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error
	Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error)
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error)
}

// subRepoPermsStore is the unified interface for managing sub repository
// permissions explicitly in the database. It is concurrency-safe and maintains
// data consistency over sub_repo_permissions table.
type subRepoPermsStore struct {
	*basestore.Store
}

// SubRepoPerms returns a new SubRepoPermsStore with the given parameters.
func SubRepoPerms(db dbutil.DB) SubRepoPermsStore {
	return &subRepoPermsStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

func SubRepoPermsWith(other basestore.ShareableStore) SubRepoPermsStore {
	return &subRepoPermsStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *subRepoPermsStore) With(other basestore.ShareableStore) SubRepoPermsStore {
	return &subRepoPermsStore{Store: s.Store.With(other)}
}

// Transact begins a new transaction and make a new SubRepoPermsStore over it.
func (s *subRepoPermsStore) Transact(ctx context.Context) (SubRepoPermsStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &subRepoPermsStore{Store: txBase}, err
}

func (s *subRepoPermsStore) Done(err error) error {
	return s.Store.Done(err)
}

// Upsert will upsert sub repo permissions data.
func (s *subRepoPermsStore) Upsert(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissions) error {
	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, path_includes, path_excludes, version, updated_at)
VALUES (%s, %s, %s, %s, %s, now())
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  path_includes = EXCLUDED.path_includes,
  path_excludes = EXCLUDED.path_excludes,
  version = EXCLUDED.version,
  updated_at = now()
`, userID, repoID, pq.Array(perms.PathIncludes), pq.Array(perms.PathExcludes), SubRepoPermsVersion)
	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions")
}

// UpsertWithSpec will upsert sub repo permissions data using the provided
// external repo spec to map to out internal repo id. If there is no mapping,
// nothing is written.
func (s *subRepoPermsStore) UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error {
	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, path_includes, path_excludes, version, updated_at)
SELECT %s, id, %s, %s, %s, now()
FROM repo
WHERE external_service_id = %s
  AND external_service_type = %s
  AND external_id = %s
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  path_includes = EXCLUDED.path_includes,
  path_excludes = EXCLUDED.path_excludes,
  version = EXCLUDED.version,
  updated_at = now()
`, userID, pq.Array(perms.PathIncludes), pq.Array(perms.PathExcludes), SubRepoPermsVersion, spec.ServiceID, spec.ServiceType, spec.ID)

	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions with spec")
}

// Get will fetch sub repo rules for the given repo and user combination.
func (s *subRepoPermsStore) Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT path_includes, path_excludes
FROM sub_repo_permissions
WHERE user_id = %s
  AND repo_id = %s
  AND version = %s
`, userID, repoID, SubRepoPermsVersion)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions")
	}

	perms := new(authz.SubRepoPermissions)
	for rows.Next() {
		var includes []string
		var excludes []string
		if err := rows.Scan(pq.Array(&includes), pq.Array(&excludes)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		perms.PathIncludes = append(perms.PathIncludes, includes...)
		perms.PathExcludes = append(perms.PathExcludes, excludes...)
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return perms, nil
}

// GetByUser fetches all sub repo perms for a user keyed by repo.
func (s *subRepoPermsStore) GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT r.name, path_includes, path_excludes
FROM sub_repo_permissions
JOIN repo r on r.id = repo_id
WHERE user_id = %s
  AND version = %s
`, userID, SubRepoPermsVersion)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions by user")
	}

	result := make(map[api.RepoName]authz.SubRepoPermissions)
	for rows.Next() {
		var perms authz.SubRepoPermissions
		var repoName api.RepoName
		if err := rows.Scan(&repoName, pq.Array(&perms.PathIncludes), pq.Array(&perms.PathExcludes)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		result[repoName] = perms
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return result, nil
}
