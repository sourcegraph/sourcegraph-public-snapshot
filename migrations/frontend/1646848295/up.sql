ALTER TABLE lsif_uploads ADD COLUMN IF NOT EXISTS indexer_version text;
COMMENT ON COLUMN lsif_uploads.indexer_version IS 'The version of the indexer that produced the index file. If not supplied by the user it will be pulled from the index metadata.';

DROP VIEW IF EXISTS lsif_uploads_with_repository_name;
DROP VIEW IF EXISTS lsif_dumps_with_repository_name;
DROP VIEW IF EXISTS lsif_dumps;

CREATE VIEW lsif_uploads_with_repository_name AS
SELECT
    u.id,
    u.commit,
    u.root,
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
    r.name AS repository_name
FROM lsif_uploads u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;

CREATE VIEW lsif_dumps AS
SELECT
    u.id,
    u.commit,
    u.root,
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
    u.finished_at AS processed_at
FROM lsif_uploads u
WHERE u.state = 'completed'::text OR u.state = 'deleting'::text;

CREATE VIEW lsif_dumps_with_repository_name AS
SELECT
    u.id,
    u.commit,
    u.root,
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
    u.processed_at,
    r.name AS repository_name
FROM lsif_dumps u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;
