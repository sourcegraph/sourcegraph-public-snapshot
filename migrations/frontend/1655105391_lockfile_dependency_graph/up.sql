DELETE FROM codeintel_lockfiles;
DELETE FROM codeintel_lockfile_references;

ALTER TABLE codeintel_lockfiles
  ADD COLUMN IF NOT EXISTS lockfile text;

-- This is a backwards incompatible change. See dev/ci/go-backcompat/flakefiles/v3.41.0.json for details.
DROP INDEX IF EXISTS codeintel_lockfiles_repository_id_commit_bytea;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfiles_repository_id_commit_bytea_lockfile ON codeintel_lockfiles USING btree (repository_id, commit_bytea, lockfile);

COMMENT ON COLUMN codeintel_lockfiles.lockfile IS 'Relative path of a lockfile in the given repository and the given commit.';

ALTER TABLE codeintel_lockfile_references
  -- We can't make them non-nullable to stay backwards compatible
  ADD COLUMN IF NOT EXISTS depends_on integer[] DEFAULT '{}',
  ADD COLUMN IF NOT EXISTS resolution_lockfile text,
  ADD COLUMN IF NOT EXISTS resolution_repository_id integer,
  ADD COLUMN IF NOT EXISTS resolution_commit_bytea bytea;

COMMENT ON COLUMN codeintel_lockfile_references.depends_on IS 'IDs of other `codeintel_lockfile_references` this package depends on in the context of this `codeintel_lockfile_references.resolution_id`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_lockfile IS 'Relative path of lockfile in which this package was referenced. Corresponds to `codeintel_lockfiles.lockfile`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_repository_id IS 'ID of the repository in which lockfile was resolved. Corresponds to `codeintel_lockfiles.repository_id`.';
COMMENT ON COLUMN codeintel_lockfile_references.resolution_commit_bytea IS 'Commit at which lockfile was resolved. Corresponds to `codeintel_lockfiles.commit_bytea`.';

CREATE INDEX IF NOT EXISTS codeintel_lockfiles_references_depends_on
ON codeintel_lockfile_references USING GIN (depends_on gin__int_ops);

-- This is a backwards incompatible change. See dev/ci/go-backcompat/flakefiles/v3.41.0.json for details.
DROP INDEX IF EXISTS codeintel_lockfile_references_repository_name_revspec_package;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package_resolution ON codeintel_lockfile_references USING btree (
    repository_name,
    revspec,
    package_scheme,
    package_name,
    package_version,
    resolution_lockfile,
    resolution_repository_id,
    resolution_commit_bytea
);
