BEGIN;

ALTER TABLE lsif_index_configuration
  ADD COLUMN "autoindex_enabled" BOOLEAN NOT NULL DEFAULT TRUE;

COMMENT ON COLUMN lsif_index_configuration.autoindex_enabled IS 'Whether or not auto-indexing should be attempted on this repo. Index jobs may be inferred from the repository contents if data is empty.';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
