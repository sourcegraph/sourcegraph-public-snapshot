BEGIN;

-- Drop view and index that depends on this type
DROP VIEW lsif_dumps;
DROP INDEX lsif_uploads_repository_id_commit_root_indexer;

-- Create old enum
CREATE TYPE lsif_upload_state_temp AS ENUM(
    'queued',
    'processing',
    'completed',
    'errored'
);

-- Update type of state column
ALTER TABLE lsif_uploads
    DROP COLUMN num_parts,
    DROP COLUMN uploaded_parts,
    ALTER COLUMN state DROP DEFAULT,
    ALTER COLUMN state TYPE lsif_upload_state_temp USING (CASE state WHEN 'uploading' THEN 'queued' ELSE state::text END)::lsif_upload_state_temp,
    ALTER COLUMN state SET DEFAULT 'queued';

-- Switch enum names
DROP TYPE lsif_upload_state;
ALTER TYPE lsif_upload_state_temp RENAME TO lsif_upload_state;

-- Restore index and view
CREATE UNIQUE INDEX lsif_uploads_repository_id_commit_root_indexer ON lsif_uploads(repository_id, "commit", root, indexer) WHERE state = 'completed'::lsif_upload_state;
CREATE VIEW lsif_dumps AS SELECT u.*, u.finished_at as processed_at FROM lsif_uploads u WHERE state = 'completed';

COMMIT;
