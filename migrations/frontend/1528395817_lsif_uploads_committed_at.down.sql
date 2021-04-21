BEGIN;

DROP INDEX lsif_uploads_committed_at;
ALTER TABLE lsif_uploads DROP COLUMN committed_at;

COMMIT;
