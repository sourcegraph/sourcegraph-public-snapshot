BEGIN;

-- Drop dependent views
DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_dumps;
DROP VIEW lsif_uploads_with_repository_name;

-- Add column to rate limit scanning of individual upload records
ALTER TABLE lsif_uploads ADD COLUMN format TEXT NOT NULL DEFAULT '.lsif';
COMMENT ON COLUMN lsif_uploads.format IS 'The format of this upload (e.g. .lsif, .lsif-flat, .lsif-pb).';

-- Update view definitions to include new fields
CREATE VIEW lsif_uploads_with_repository_name AS
    SELECT u.id,
        u.commit,
        u.root,
        u.uploaded_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.indexer,
        u.num_parts,
        u.uploaded_parts,
        u.process_after,
        u.num_resets,
        u.upload_size,
        u.num_failures,
        u.associated_index_id,
        u.expired,
        u.last_retention_scan_at,
        u.format,
        r.name AS repository_name
    FROM lsif_uploads u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_dumps AS
    SELECT u.id,
        u.commit,
        u.root,
        u.uploaded_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.indexer,
        u.num_parts,
        u.uploaded_parts,
        u.process_after,
        u.num_resets,
        u.upload_size,
        u.num_failures,
        u.associated_index_id,
        u.expired,
        u.last_retention_scan_at,
        u.finished_at AS processed_at,
        u.format
    FROM lsif_uploads u
    WHERE u.state = 'completed'::text OR u.state = 'deleting'::text;

CREATE VIEW lsif_dumps_with_repository_name AS
    SELECT u.id,
        u.commit,
        u.root,
        u.uploaded_at,
        u.state,
        u.failure_message,
        u.started_at,
        u.finished_at,
        u.repository_id,
        u.indexer,
        u.num_parts,
        u.uploaded_parts,
        u.process_after,
        u.num_resets,
        u.upload_size,
        u.num_failures,
        u.associated_index_id,
        u.expired,
        u.last_retention_scan_at,
        u.processed_at,
        u.format,
        r.name AS repository_name
    FROM lsif_dumps u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;
