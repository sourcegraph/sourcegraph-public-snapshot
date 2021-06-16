BEGIN;

ALTER TABLE lsif_index_configuration
  ADD COLUMN "autoindex_enabled" BOOLEAN NOT NULL DEFAULT TRUE;

-- TODO: Later we should remove lsif_indexable_repositories

COMMIT;
