package stores

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jackc/pgconn"
	"github.com/keegancsmith/sqlf"
	"github.com/lib/pq"

	"github.com/sourcegraph/sourcegraph/internal/database/basestore"
	"github.com/sourcegraph/sourcegraph/internal/database/dbutil"
	"github.com/sourcegraph/sourcegraph/lib/errors"
)

// Release describes a release of an extension in the extension registry.
type Release struct {
	ID                  int64
	RegistryExtensionID int32
	CreatorUserID       int32
	ReleaseVersion      *string
	ReleaseTag          string
	Manifest            string
	Bundle              *string
	SourceMap           *string
	CreatedAt           time.Time
}

// ReleaseNotFoundError occurs when an extension release is not found in the
// extension registry.
type ReleaseNotFoundError struct {
	args []any
}

// NotFound implements errcode.NotFounder.
func (err ReleaseNotFoundError) NotFound() bool { return true }

func (err ReleaseNotFoundError) Error() string {
	return fmt.Sprintf("registry extension release not found: %v", err.args)
}

var ErrInvalidJSONInManifest = errors.New("invalid syntax in extension manifest JSON")

type ReleaseStore interface {
	// Create creates a new release of an extension in the extension registry. The release.ID and
	// release.CreatedAt fields are ignored (they are populated automatically by the database).
	Create(ctx context.Context, release *Release) (id int64, err error)
	// GetLatest gets the latest release for the extension with the given release tag (e.g., "release").
	// If includeArtifacts is true, it populates the (*dbRelease).{Bundle,SourceMap} fields, which may be large.
	GetLatest(ctx context.Context, registryExtensionID int32, releaseTag string, includeArtifacts bool) (*Release, error)
	// GetLatestBatch gets the latest releases for the extensions with the given release tag
	// (e.g., "release"). If includeArtifacts is true, it populates the (*dbRelease).{Bundle,SourceMap}
	// fields, which may be large.
	GetLatestBatch(ctx context.Context, registryExtensionIDs []int32, releaseTag string, includeArtifacts bool) ([]*Release, error)
	// GetArtifacts gets the bundled JavaScript source file contents and the source map for a release
	// (by ID).
	GetArtifacts(ctx context.Context, id int64) (bundle, sourcemap []byte, err error)

	Transact(context.Context) (ReleaseStore, error)
	With(basestore.ShareableStore) ReleaseStore
	basestore.ShareableStore
}

type releaseStore struct {
	*basestore.Store
}

var _ ReleaseStore = (*releaseStore)(nil)

// Releases instantiates and returns a new ReleasesStore with prepared statements.
func Releases(db dbutil.DB) ReleaseStore {
	return &releaseStore{Store: basestore.NewWithDB(db, sql.TxOptions{})}
}

// ReleasesWith instantiates and returns a new ReleasesStore using the other store handle.
func ReleasesWith(other basestore.ShareableStore) ReleaseStore {
	return &releaseStore{Store: basestore.NewWithHandle(other.Handle())}
}

func (s *releaseStore) With(other basestore.ShareableStore) ReleaseStore {
	return &releaseStore{Store: s.Store.With(other)}
}

func (s *releaseStore) Transact(ctx context.Context) (ReleaseStore, error) {
	txBase, err := s.Store.Transact(ctx)
	return &releaseStore{Store: txBase}, err
}

func (s *releaseStore) Create(ctx context.Context, release *Release) (id int64, err error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/releases.go:Create
INSERT INTO registry_extension_releases
	(registry_extension_id, creator_user_id, release_version, release_tag, manifest, bundle, source_map)
VALUES
	(%s, %s, %s, %s, %s, %s, %s)
RETURNING id
`,
		release.RegistryExtensionID,
		release.CreatorUserID,
		release.ReleaseVersion,
		release.ReleaseTag,
		release.Manifest,
		release.Bundle,
		release.SourceMap,
	)

	if err := s.QueryRow(ctx, q).Scan(&id); err != nil {
		var e *pgconn.PgError
		if errors.As(err, &e) && e.Message == "invalid input syntax for type json" {
			return 0, ErrInvalidJSONInManifest
		}
		return 0, err
	}

	return id, nil
}

func (s *releaseStore) GetLatest(ctx context.Context, registryExtensionID int32, releaseTag string, includeArtifacts bool) (*Release, error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/releases.go:GetLatest
SELECT
	id,
	registry_extension_id,
	creator_user_id,
	release_version,
	release_tag,
	manifest,
	CASE WHEN %v::boolean THEN bundle ELSE null END AS bundle,
	CASE WHEN %v::boolean THEN source_map ELSE null END AS source_map,
	created_at
FROM
	registry_extension_releases
WHERE
	registry_extension_id = %d
	AND
	release_tag = %s
	AND
	deleted_at IS NULL
ORDER BY
	created_at DESC
LIMIT 1`,
		includeArtifacts,
		includeArtifacts,
		registryExtensionID,
		releaseTag,
	)
	var r Release
	err := s.QueryRow(ctx, q).Scan(&r.ID, &r.RegistryExtensionID, &r.CreatorUserID, &r.ReleaseVersion, &r.ReleaseTag, &r.Manifest, &r.Bundle, &r.SourceMap, &r.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ReleaseNotFoundError{[]any{fmt.Sprintf("latest for registry extension ID %d tag %q", registryExtensionID, releaseTag)}}
		}
		return nil, err
	}
	return &r, nil
}

func (s *releaseStore) GetLatestBatch(ctx context.Context, registryExtensionIDs []int32, releaseTag string, includeArtifacts bool) ([]*Release, error) {
	if len(registryExtensionIDs) == 0 {
		return nil, nil
	}

	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/releases.go:GetLatestBatch
SELECT DISTINCT ON (rer.registry_extension_id)
	rer.id,
	rer.registry_extension_id,
	rer.creator_user_id,
	rer.release_version,
	rer.release_tag,
	rer.manifest,
	CASE WHEN %v::boolean THEN rer.bundle ELSE null END AS bundle,
	CASE WHEN %v::boolean THEN rer.source_map ELSE null END AS source_map,
	rer.created_at
FROM
	registry_extension_releases rer
WHERE
	rer.registry_extension_id = ANY (%s)
	AND
	rer.release_tag = %s
	AND rer.deleted_at IS NULL
ORDER BY
	rer.registry_extension_id,
	rer.created_at DESC
`,
		includeArtifacts,
		includeArtifacts,
		pq.Array(registryExtensionIDs),
		releaseTag,
	)

	rows, err := s.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer func() { err = basestore.CloseRows(rows, err) }()

	var releases []*Release
	for rows.Next() {
		var r Release
		err := rows.Scan(&r.ID, &r.RegistryExtensionID, &r.CreatorUserID, &r.ReleaseVersion, &r.ReleaseTag, &r.Manifest, &r.Bundle, &r.SourceMap, &r.CreatedAt)
		if err != nil {
			return nil, err
		}
		releases = append(releases, &r)
	}

	return releases, nil
}

func (s *releaseStore) GetArtifacts(ctx context.Context, id int64) (bundle, sourcemap []byte, err error) {
	q := sqlf.Sprintf(`
-- source: enterprise/cmd/frontend/internal/registry/stores/releases.go:GetArtifacts
SELECT
	bundle,
	source_map
FROM
	registry_extension_releases
WHERE
	id = %d
	AND
	deleted_at IS NULL`,
		id,
	)

	if err := s.QueryRow(ctx, q).Scan(&bundle, &sourcemap); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, ReleaseNotFoundError{[]any{fmt.Sprintf("registry extension release %d", id)}}
		}
		return nil, nil, err
	}

	if bundle == nil {
		return nil, nil, ReleaseNotFoundError{[]any{fmt.Sprintf("no bundle for registry extension release %d", id)}}
	}

	return bundle, sourcemap, nil
}
