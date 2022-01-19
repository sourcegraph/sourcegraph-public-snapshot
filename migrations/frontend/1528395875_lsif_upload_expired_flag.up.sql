-- +++
-- parent: 1528395874
-- +++

BEGIN;

-- Drop dependent views
DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_dumps;
DROP VIEW lsif_uploads_with_repository_name;

-- Add new column
ALTER TABLE lsif_uploads ADD COLUMN expired boolean not null default false;
COMMENT ON COLUMN lsif_uploads.expired IS 'Whether or not this upload data is no longer protected by any data retention policy.';

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
        u.finished_at AS processed_at
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
        u.processed_at,
        r.name AS repository_name
    FROM lsif_dumps u
    JOIN repo r ON r.id = u.repository_id
    WHERE r.deleted_at IS NULL;

COMMIT;
