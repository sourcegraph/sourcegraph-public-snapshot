package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/keegancsmith/sqlf"
	"github.com/sourcegraph/sourcegraph/cmd/frontend/db/dbconn"
)

// RegistryExtensionRelease describes a release of an extension in the extension registry.
type RegistryExtensionRelease struct {
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

type registryExtensionReleases struct{}

// RegistryExtensionReleaseNotFoundError occurs when an extension release is not found in the
// extension registry.
type RegistryExtensionReleaseNotFoundError struct {
	args []interface{}
}

// NotFound implements errcode.NotFounder.
func (err RegistryExtensionReleaseNotFoundError) NotFound() bool { return true }

func (err RegistryExtensionReleaseNotFoundError) Error() string {
	return fmt.Sprintf("registry extension release not found: %v", err.args)
}

// Create creates a new release of an extension in the extension registry. The release.ID and
// release.CreatedAt fields are ignored (they are populated automatically by the database).
func (s *registryExtensionReleases) Create(ctx context.Context, release *RegistryExtensionRelease) (id int64, err error) {
	if Mocks.RegistryExtensionReleases.Create != nil {
		return Mocks.RegistryExtensionReleases.Create(release)
	}

	if err := dbconn.Global.QueryRowContext(ctx,
		`
INSERT INTO registry_extension_releases(registry_extension_id, creator_user_id, release_version, release_tag, manifest, bundle, source_map)
VALUES($1, $2, $3, $4, $5, $6, $7)
RETURNING id
`,
		release.RegistryExtensionID, release.CreatorUserID, release.ReleaseVersion, release.ReleaseTag, release.Manifest, release.Bundle, release.SourceMap,
	).Scan(&id); err != nil {
		return 0, err
	}
	return id, nil
}

// GetLatest gets the latest release for the extension with the given release tag (e.g.,
// "release"). If includeArtifacts is true, it populates the
// (*RegistryExtensionRelease).{Bundle,SourceMap} fields, which may be large.
func (s *registryExtensionReleases) GetLatest(ctx context.Context, registryExtensionID int32, releaseTag string, includeArtifacts bool) (*RegistryExtensionRelease, error) {
	if Mocks.RegistryExtensionReleases.GetLatest != nil {
		return Mocks.RegistryExtensionReleases.GetLatest(registryExtensionID, releaseTag, includeArtifacts)
	}

	q := sqlf.Sprintf(`
SELECT id, registry_extension_id, creator_user_id, release_version, release_tag, manifest, CASE WHEN %v::boolean THEN bundle ELSE null END AS bundle, CASE WHEN %v::boolean THEN source_map ELSE null END AS source_map, created_at
FROM registry_extension_releases
WHERE registry_extension_id=%d AND release_tag=%s AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 1`, includeArtifacts, includeArtifacts, registryExtensionID, releaseTag)
	var r RegistryExtensionRelease
	err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&r.ID, &r.RegistryExtensionID, &r.CreatorUserID, &r.ReleaseVersion, &r.ReleaseTag, &r.Manifest, &r.Bundle, &r.SourceMap, &r.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, RegistryExtensionReleaseNotFoundError{[]interface{}{fmt.Sprintf("latest for registry extension ID %d tag %q", registryExtensionID, releaseTag)}}
		}
		return nil, err
	}
	return &r, nil
}

// GetArtifacts gets the bundled JavaScript source file contents and the source map for a release
// (by ID).
func (s *registryExtensionReleases) GetArtifacts(ctx context.Context, id int64) (bundle, sourcemap []byte, err error) {
	q := sqlf.Sprintf(`
SELECT bundle, source_map
FROM registry_extension_releases
WHERE id=%d AND deleted_at IS NULL`, id)
	if err := dbconn.Global.QueryRowContext(ctx, q.Query(sqlf.PostgresBindVar), q.Args()...).Scan(&bundle, &sourcemap); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, RegistryExtensionReleaseNotFoundError{[]interface{}{fmt.Sprintf("registry extension release %d", id)}}
		}
		return nil, nil, err
	}
	if bundle == nil {
		return nil, nil, RegistryExtensionReleaseNotFoundError{[]interface{}{fmt.Sprintf("no bundle for registry extension release %d", id)}}
	}
	return bundle, sourcemap, nil
}

// MockRegistryExtensionReleases mocks the registry extension releases store.
type MockRegistryExtensionReleases struct {
	Create    func(release *RegistryExtensionRelease) (int64, error)
	GetLatest func(registryExtensionID int32, releaseTag string, includeArtifacts bool) (*RegistryExtensionRelease, error)
}
