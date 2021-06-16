BEGIN;

ALTER TABLE lsif_index_configuration
  ADD COLUMN "autoindex_enabled" BOOLEAN NOT NULL DEFAULT TRUE;

COMMIT;
