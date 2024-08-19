package database

import (
	"context"

	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/api"
	"github.com/sourcegraph/sourcegraph/internal/authz"
	"github.com/sourcegraph/sourcegraph/internal/conf"
	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/extsvc"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// SubRepoPermsVersion is defines the version we are using to encode our include
// and exclude patterns.
const SubRepoPermsVersion = 1

type SubRepoPermsStore interface {
	basestore.ShareableStore
	With(other basestore.ShareableStore) SubRepoPermsStore
	Transact(ctx context.Context) (SubRepoPermsStore, error)
	Done(err error) error
	Upsert(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissions) error
	UpsertWithIPs(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissionsWithIPs) error
	UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error
	UpsertWithSpecWithIPs(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissionsWithIPs) error
	Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error)
	GetWithIPs(ctx context.Context, userID int32, repoID api.RepoID, backfillWithWildcardIP bool) (*authz.SubRepoPermissionsWithIPs, error)
	GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error)
	GetByUserWithIPs(ctx context.Context, userID int32, backfillWithWildcardIP bool) (map[api.RepoName]authz.SubRepoPermissionsWithIPs, error)
	// GetByUserAndService gets the sub repo permissions for a user, but filters down
	// to only repos that come from a specific external service.
	GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (map[api.ExternalRepoSpec]authz.SubRepoPermissions, error)
	GetByUserAndServiceWithIPs(ctx context.Context, userID int32, serviceType string, serviceID string, backfillWithWildcardIP bool) (map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs, error)
	RepoIDSupported(ctx context.Context, repoID api.RepoID) (bool, error)
	RepoSupported(ctx context.Context, repo api.RepoName) (bool, error)
	DeleteByUser(ctx context.Context, userID int32) error
}

// subRepoPermsStore is the unified interface for managing sub repository
// permissions explicitly in the database. It is concurrency-safe and maintains
// data consistency over sub_repo_permissions table.
type subRepoPermsStore struct {
	*basestore.Store
}

var _ SubRepoPermsStore = (*subRepoPermsStore)(nil)

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
INSERT INTO sub_repo_permissions (user_id, repo_id, paths, version, updated_at, ips)
VALUES (%s, %s, %s, %s, now(), NULL)
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  paths = EXCLUDED.paths,
  version = EXCLUDED.version,
  updated_at = now(),
  ips = NULL
`, userID, repoID, pq.Array(perms.Paths), SubRepoPermsVersion)
	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions")
}

// UpsertWithIPs will upsert sub repo permissions data.
func (s *subRepoPermsStore) UpsertWithIPs(ctx context.Context, userID int32, repoID api.RepoID, perms authz.SubRepoPermissionsWithIPs) error {
	var paths []string
	var ips []string

	for _, p := range perms.Paths {
		paths = append(paths, p.Path)
		ips = append(ips, p.IP)
	}

	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, paths, version, updated_at, ips)
	VALUES (%s, %s, %s, %s, now(), %s)
	ON CONFLICT (user_id, repo_id, version)
		DO UPDATE
		SET
			user_id = EXCLUDED.user_ID,
			repo_id = EXCLUDED.repo_id,
			paths = EXCLUDED.paths,
			version = EXCLUDED.version,
			updated_at = now(),
			ips = EXCLUDED.ips
`, userID, repoID, pq.Array(paths), SubRepoPermsVersion, pq.Array(ips))

	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions with IP")
}

// UpsertWithSpec will upsert sub repo permissions data using the provided
// external repo spec to map to our internal repo id. If there is no mapping,
// nothing is written.
func (s *subRepoPermsStore) UpsertWithSpec(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissions) error {
	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, paths, version, updated_at)
SELECT %s, id, %s, %s, now()
FROM repo
WHERE external_service_id = %s
  AND external_service_type = %s
  AND external_id = %s
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  paths = EXCLUDED.paths,
  version = EXCLUDED.version,
  updated_at = now(),
  ips = NULL
`, userID, pq.Array(perms.Paths), SubRepoPermsVersion, spec.ServiceID, spec.ServiceType, spec.ID)

	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions with spec")
}

// UpsertWithSpecWithIPs will upsert sub repo permissions data using the provided
// external repo spec to map to our internal repo id. If there is no mapping,
// nothing is written.
func (s *subRepoPermsStore) UpsertWithSpecWithIPs(ctx context.Context, userID int32, spec api.ExternalRepoSpec, perms authz.SubRepoPermissionsWithIPs) error {
	var paths []string
	var ips []string

	for _, p := range perms.Paths {
		paths = append(paths, p.Path)
		ips = append(ips, p.IP)
	}

	q := sqlf.Sprintf(`
INSERT INTO sub_repo_permissions (user_id, repo_id, paths, version, updated_at, ips)
SELECT %s, id, %s, %s, now(), %s
FROM repo
WHERE external_service_id = %s
  AND external_service_type = %s
  AND external_id = %s
ON CONFLICT (user_id, repo_id, version)
DO UPDATE
SET
  user_id = EXCLUDED.user_ID,
  repo_id = EXCLUDED.repo_id,
  paths = EXCLUDED.paths,
  version = EXCLUDED.version,
  updated_at = now(),
  ips = EXCLUDED.ips
`, userID, pq.Array(paths), SubRepoPermsVersion, pq.Array(ips), spec.ServiceID, spec.ServiceType, spec.ID)

	return errors.Wrap(s.Exec(ctx, q), "upserting sub repo permissions with spec and IPs")
}

// GetWithIPs will fetch sub repo rules for the given repo and user combination.
func (s *subRepoPermsStore) GetWithIPs(ctx context.Context, userID int32, repoID api.RepoID, backfillWithWildcardIP bool) (*authz.SubRepoPermissionsWithIPs, error) {
	q := sqlf.Sprintf(`
SELECT paths, ips
FROM sub_repo_permissions
WHERE repo_id = %s
  AND user_id = %s
  AND version = %s
  `, userID, repoID, SubRepoPermsVersion)
	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions")
	}

	perms := new(authz.SubRepoPermissionsWithIPs)
	for rows.Next() {
		var paths []string
		var ips []string
		if err := rows.Scan(pq.Array(&paths), pq.Array(&ips)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}

		if len(paths) != len(ips) {
			if len(ips) != 0 {
				return nil, errors.Errorf("database constraint violation for repo_id=%d, user_id=%d, version=%d: len(paths)!= len(ips)", repoID, userID, SubRepoPermsVersion)
			}

			// The 'ips' column is empty
			if !backfillWithWildcardIP {
				return nil, errors.Wrapf(IPsNotSyncedError, "repo_id=%d, user_id=%d, version=%d", repoID, userID, SubRepoPermsVersion)
			}

			// Backfill each path with a wildcard IP
			ips = make([]string, len(paths))
			for i := 0; i < len(paths); i++ {
				ips[i] = "*"
			}
		}

		for i := 0; i < len(paths); i++ {
			perms.Paths = append(perms.Paths, authz.PathWithIP{Path: paths[i], IP: ips[i]})
		}
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return perms, nil
}

// Get will fetch sub repo rules for the given repo and user combination.
func (s *subRepoPermsStore) Get(ctx context.Context, userID int32, repoID api.RepoID) (*authz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT paths
FROM sub_repo_permissions
WHERE repo_id = %s
  AND user_id = %s
  AND version = %s
`, userID, repoID, SubRepoPermsVersion)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions")
	}

	perms := new(authz.SubRepoPermissions)
	for rows.Next() {
		var paths []string
		if err := rows.Scan(pq.Array(&paths)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		perms.Paths = append(perms.Paths, paths...)
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return perms, nil
}

// GetByUser fetches all sub repo perms for a user keyed by repo.
func (s *subRepoPermsStore) GetByUser(ctx context.Context, userID int32) (map[api.RepoName]authz.SubRepoPermissions, error) {
	enforceForSiteAdmins := conf.Get().AuthzEnforceForSiteAdmins

	q := sqlf.Sprintf(`
	SELECT r.name, paths
	FROM sub_repo_permissions
	JOIN repo r on r.id = repo_id
	JOIN users u on u.id = user_id
	WHERE user_id = %s
	AND version = %s
	-- When user is a site admin and AuthzEnforceForSiteAdmins is FALSE
	-- we want to return zero results. This causes us to fall back to
	-- repo level checks and allows access to all paths in all repos.
	AND NOT (u.site_admin AND NOT %t)
	`, userID, SubRepoPermsVersion, enforceForSiteAdmins)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions by user")
	}

	result := make(map[api.RepoName]authz.SubRepoPermissions)
	for rows.Next() {
		var perms authz.SubRepoPermissions
		var repoName api.RepoName
		if err := rows.Scan(&repoName, pq.Array(&perms.Paths)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		result[repoName] = perms
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return result, nil
}

// GetByUserWithIPs fetches all sub repo perms for a user keyed by repo.
func (s *subRepoPermsStore) GetByUserWithIPs(ctx context.Context, userID int32, backfillWithWildcardIP bool) (map[api.RepoName]authz.SubRepoPermissionsWithIPs, error) {
	enforceForSiteAdmins := conf.Get().AuthzEnforceForSiteAdmins

	q := sqlf.Sprintf(`
    SELECT r.name, paths, ips
    FROM sub_repo_permissions
    JOIN repo r on r.id = repo_id
    JOIN users u on u.id = user_id
    WHERE user_id = %s
    AND version = %s
    -- When user is a site admin and AuthzEnforceForSiteAdmins is FALSE
	-- we want to return zero results. This causes us to fall back to
	-- repo level checks and allows access to all paths in all repos.
    AND NOT (u.site_admin AND NOT %t)
    `, userID, SubRepoPermsVersion, enforceForSiteAdmins)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions by user with IPs")
	}
	defer rows.Close()

	result := make(map[api.RepoName]authz.SubRepoPermissionsWithIPs)
	for rows.Next() {
		var repoName api.RepoName
		var paths, ips []string
		if err := rows.Scan(&repoName, pq.Array(&paths), pq.Array(&ips)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}

		perms := authz.SubRepoPermissionsWithIPs{}
		if len(paths) != len(ips) {
			if len(ips) != 0 {
				return nil, errors.Errorf("database constraint violation for user_id=%d, repo=%s: len(paths) != len(ips)", userID, repoName)
			}

			if !backfillWithWildcardIP {
				return nil, errors.Wrapf(IPsNotSyncedError, "user_id=%d, repo=%s", userID, repoName)
			}

			ips = make([]string, len(paths))
			for i := range paths {
				ips[i] = "*"
			}
		}

		for i := range paths {
			perms.Paths = append(perms.Paths, authz.PathWithIP{Path: paths[i], IP: ips[i]})
		}

		result[repoName] = perms
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating rows")
	}

	return result, nil
}

func (s *subRepoPermsStore) GetByUserAndService(ctx context.Context, userID int32, serviceType string, serviceID string) (map[api.ExternalRepoSpec]authz.SubRepoPermissions, error) {
	q := sqlf.Sprintf(`
SELECT r.external_id, paths
FROM sub_repo_permissions
JOIN repo r on r.id = repo_id
WHERE user_id = %s
  AND version = %s
  AND r.external_service_type = %s
  AND r.external_service_id = %s
`, userID, SubRepoPermsVersion, serviceType, serviceID)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions by user and service")
	}

	result := make(map[api.ExternalRepoSpec]authz.SubRepoPermissions)
	for rows.Next() {
		var perms authz.SubRepoPermissions
		spec := api.ExternalRepoSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
		}
		if err := rows.Scan(&spec.ID, pq.Array(&perms.Paths)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}
		result[spec] = perms
	}

	if err := rows.Close(); err != nil {
		return nil, errors.Wrap(err, "closing rows")
	}

	return result, nil
}

func (s *subRepoPermsStore) GetByUserAndServiceWithIPs(ctx context.Context, userID int32, serviceType string, serviceID string, backfillWithWildcardIP bool) (map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs, error) {
	q := sqlf.Sprintf(`
SELECT r.external_id, paths, ips
FROM sub_repo_permissions
JOIN repo r on r.id = repo_id
WHERE user_id = %s
  AND version = %s
  AND r.external_service_type = %s
  AND r.external_service_id = %s
`, userID, SubRepoPermsVersion, serviceType, serviceID)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, errors.Wrap(err, "getting sub repo permissions by user and service with IPs")
	}
	defer rows.Close()

	result := make(map[api.ExternalRepoSpec]authz.SubRepoPermissionsWithIPs)
	for rows.Next() {
		var externalID string
		var paths, ips []string
		if err := rows.Scan(&externalID, pq.Array(&paths), pq.Array(&ips)); err != nil {
			return nil, errors.Wrap(err, "scanning row")
		}

		spec := api.ExternalRepoSpec{
			ServiceType: serviceType,
			ServiceID:   serviceID,
			ID:          externalID,
		}

		perms := authz.SubRepoPermissionsWithIPs{}
		if len(paths) != len(ips) {
			if len(ips) != 0 {
				return nil, errors.Errorf("database constraint violation for user_id=%d, external_id=%s: len(paths) != len(ips)", userID, externalID)
			}

			if !backfillWithWildcardIP {
				return nil, errors.Wrapf(IPsNotSyncedError, "user_id=%d, external_id=%s", userID, externalID)
			}

			ips = make([]string, len(paths))
			for i := range paths {
				ips[i] = "*"
			}
		}

		for i := range paths {
			perms.Paths = append(perms.Paths, authz.PathWithIP{Path: paths[i], IP: ips[i]})
		}

		result[spec] = perms
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "iterating rows")
	}

	return result, nil
}

// RepoIDSupported returns true if repo with the given ID has sub-repo permissions
// (i.e. it is private and its type is one of the SubRepoSupportedCodeHostTypes)
func (s *subRepoPermsStore) RepoIDSupported(ctx context.Context, repoID api.RepoID) (bool, error) {
	q := sqlf.Sprintf(`
SELECT EXISTS(
SELECT
FROM repo
WHERE id = %s
AND private = TRUE
AND external_service_type = %s
)
`, repoID, extsvc.TypePerforce)

	exists, _, err := basestore.ScanFirstBool(s.Query(ctx, q))
	if err != nil {
		return false, errors.Wrap(err, "querying database")
	}
	return exists, nil
}

// RepoSupported returns true if repo has sub-repo permissions
// (i.e. it is private and its type is one of the SubRepoSupportedCodeHostTypes)
func (s *subRepoPermsStore) RepoSupported(ctx context.Context, repo api.RepoName) (bool, error) {
	q := sqlf.Sprintf(`
SELECT EXISTS(
SELECT
FROM repo
WHERE name = %s
AND private = TRUE
AND external_service_type = %s
)
`, repo, extsvc.TypePerforce)

	exists, _, err := basestore.ScanFirstBool(s.Query(ctx, q))
	if err != nil {
		return false, errors.Wrap(err, "querying database")
	}
	return exists, nil
}

// DeleteByUser deletes all rows associated with the given user
func (s *subRepoPermsStore) DeleteByUser(ctx context.Context, userID int32) error {
	q := sqlf.Sprintf(`
DELETE FROM sub_repo_permissions WHERE user_id = %d
`, userID)
	return s.Exec(ctx, q)
}

// IPsNotSyncedError is returned when the IP addresses associated with a sub-repository-permissions entry have not been synced yet.
//
// Currently, this means that the permissions provider for the given service has not been updated to include the IP addresses with each
// entry it inserts.
var IPsNotSyncedError = errors.New("the IP addresses associated with this sub-repository-permissions entry have not been synced yet")
