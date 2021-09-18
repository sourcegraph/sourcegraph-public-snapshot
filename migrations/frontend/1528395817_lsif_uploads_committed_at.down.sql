BEGIN;

DROP INDEX lsif_uploads_committed_at;
ALTER TABLE lsif_uploads DROP COLUMN committed_at;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
