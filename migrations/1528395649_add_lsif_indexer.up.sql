BEGIN;

ALTER TABLE lsif_uploads ADD COLUMN indexer TEXT;
UPDATE lsif_uploads SET indexer = '';
ALTER TABLE lsif_uploads ALTER COLUMN indexer SET NOT NULL;

-- Recreate view with new column names
DROP VIEW lsif_dumps;
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

-- Modify unique index to include indexer field
DROP INDEX lsif_uploads_repository_id_commit_root;
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root_indexer ON lsif_uploads(repository_id, "commit", root, indexer) WHERE state = 'completed'::lsif_upload_state;

COMMIT;
