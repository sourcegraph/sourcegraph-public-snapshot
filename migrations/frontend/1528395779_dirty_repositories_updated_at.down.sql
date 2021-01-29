BEGIN;

ALTER TABLE lsif_dirty_repositories DROP COLUMN updated_at;

COMMIT;
