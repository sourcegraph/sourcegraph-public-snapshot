BEGIN;

DROP INDEX lsif_uploads_commit_last_checked_at;
ALTER TABLE lsif_uploads DROP COLUMN commit_last_checked_at;
DROP INDEX lsif_indexes_commit_last_checked_at;
ALTER TABLE lsif_indexes DROP COLUMN commit_last_checked_at;

COMMIT;
