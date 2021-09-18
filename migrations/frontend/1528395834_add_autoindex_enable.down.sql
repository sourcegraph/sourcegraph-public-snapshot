BEGIN;

ALTER TABLE lsif_index_configuration
  DROP COLUMN IF EXISTS "autoindex_enabled";

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
