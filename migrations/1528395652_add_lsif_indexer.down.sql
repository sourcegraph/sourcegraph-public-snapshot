BEGIN;

-- Restore old unique index
DROP INDEX lsif_uploads_repository_id_commit_root_indexer;
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root ON lsif_uploads(repository_id, "commit", root) WHERE state = 'completed'::lsif_upload_state;

-- Drop view dependent on new column
DROP VIEW lsif_dumps;

-- Drop new column
ALTER TABLE lsif_uploads DROP COLUMN indexer;

-- Recreate view with new column names
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
