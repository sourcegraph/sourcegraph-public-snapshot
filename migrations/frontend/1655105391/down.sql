ALTER TABLE codeintel_lockfiles
  DROP COLUMN IF EXISTS resolution_id text;

ALTER TABLE codeintel_lockfile_references
  DROP COLUMN IF EXISTS depends_on integer[] NOT NULL,
  DROP COLUMN IF EXISTS resolution_id text NOT NULL;
