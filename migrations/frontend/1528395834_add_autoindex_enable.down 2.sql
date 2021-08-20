BEGIN;

ALTER TABLE lsif_index_configuration
  DROP COLUMN IF EXISTS "autoindex_enabled";

COMMIT;
