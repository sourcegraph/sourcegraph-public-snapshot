ALTER TABLE lsif_uploads
ADD COLUMN IF NOT EXISTS uncompressed_size bigint;

DROP VIEW IF EXISTS lsif_uploads_with_repository_name;

CREATE VIEW lsif_uploads_with_repository_name AS
SELECT u.id,
    u.commit,
    u.root,
    u.queued_at,
    u.uploaded_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.indexer,
    u.indexer_version,
    u.num_parts,
    u.uploaded_parts,
    u.process_after,
    u.num_resets,
    u.upload_size,
    u.num_failures,
    u.associated_index_id,
    u.expired,
    u.last_retention_scan_at,
    r.name AS repository_name,
    u.uncompressed_size
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;
