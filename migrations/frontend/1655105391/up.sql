DELETE FROM codeintel_lockfiles;
DELETE FROM codeintel_lockfile_references;

ALTER TABLE codeintel_lockfiles
  ADD COLUMN IF NOT EXISTS resolution_id text;

COMMENT ON COLUMN codeintel_lockfiles.resolution_id IS 'Unique identifier for the resolution of a lockfile in the given repository and the given commit. Correponds to `codeintel_lockfile_references.resolution_id`.';

ALTER TABLE codeintel_lockfile_references
  ADD COLUMN IF NOT EXISTS depends_on integer[] NOT NULL,
  ADD COLUMN IF NOT EXISTS resolution_id text NOT NULL;

COMMENT ON COLUMN codeintel_lockfile_references.depends_on IS 'IDs of other `codeintel_lockfile_references` this package depends on in the context of this `codeintel_lockfile_references.resolution_id`.';

CREATE INDEX IF NOT EXISTS codeintel_lockfiles_references_depends_on
ON codeintel_lockfile_references USING GIN (depends_on gin__int_ops);

DROP INDEX IF EXISTS codeintel_lockfile_references_repository_name_revspec_package;
CREATE UNIQUE INDEX IF NOT EXISTS codeintel_lockfile_references_repository_name_revspec_package_resolution ON codeintel_lockfile_references USING btree (
    repository_name,
    revspec,
    package_scheme,
    package_name,
    package_version,
    resolution_id
);
