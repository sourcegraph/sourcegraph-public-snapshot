ALTER TABLE codeintel_lockfiles
  DROP COLUMN IF EXISTS resolution_id text;

ALTER TABLE codeintel_lockfile_references
  DROP COLUMN IF EXISTS depends_on integer[] NOT NULL,
  DROP COLUMN IF EXISTS resolution_id text NOT NULL;

DROP INDEX IF EXISTS codeintel_lockfile_references_repository_name_revspec_package_resolution;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package ON codeintel_lockfile_references USING btree (
    repository_name,
    revspec,
    package_scheme,
    package_name,
    package_version
);
