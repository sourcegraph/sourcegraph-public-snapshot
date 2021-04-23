BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN commit_last_checked_at timestamp with time zone;
CREATE INDEX lsif_uploads_commit_last_checked_at ON lsif_uploads (commit_last_checked_at) WHERE state != 'deleted';
ALTER TABLE lsif_indexes ADD COLUMN commit_last_checked_at timestamp with time zone;
CREATE INDEX lsif_indexes_commit_last_checked_at ON lsif_indexes (commit_last_checked_at) WHERE state != 'deleted';

COMMIT;
