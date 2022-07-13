ALTER TABLE codeintel_lockfiles
  DROP COLUMN IF EXISTS lockfile;

DROP INDEX IF EXISTS codeintel_lockfiles_repository_id_commit_bytea_lockfile;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfiles_repository_id_commit_bytea ON codeintel_lockfiles USING btree (repository_id, commit_bytea);

ALTER TABLE codeintel_lockfile_references
  DROP COLUMN IF EXISTS depends_on,
  DROP COLUMN IF EXISTS resolution_lockfile,
  DROP COLUMN IF EXISTS resolution_repository_id,
  DROP COLUMN IF EXISTS resolution_commit_bytea;

DROP INDEX IF EXISTS codeintel_lockfile_references_repository_name_revspec_package_resolution;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package ON codeintel_lockfile_references USING btree (
    repository_name,
    revspec,
    package_scheme,
    package_name,
    package_version
);
