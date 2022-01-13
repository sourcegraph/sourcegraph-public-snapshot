-- +++
-- parent: 1528395875
-- +++

BEGIN;

-- Drop dependent views
DROP VIEW lsif_dumps_with_repository_name;
DROP VIEW lsif_dumps;
DROP VIEW lsif_uploads_with_repository_name;

-- Create table to rate limit data retention scans of a repository
CREATE TABLE lsif_last_retention_scan (
    repository_id int NOT NULL,
    last_retention_scan_at timestamp with time zone NOT NULL,

    PRIMARY KEY(repository_id)
);
COMMENT ON TABLE lsif_last_retention_scan IS 'Tracks the last time uploads a repository were checked against data retention policies.';
COMMENT ON COLUMN lsif_last_retention_scan.last_retention_scan_at IS 'The last time uploads of this repository were checked against data retention policies.';

-- Add column to rate limit scanning of individual upload records
ALTER TABLE lsif_uploads ADD COLUMN last_retention_scan_at timestamp with time zone;
COMMENT ON COLUMN lsif_uploads.last_retention_scan_at IS 'The last time this upload was checked against data retention policies.';

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
