DROP VIEW IF EXISTS lsif_uploads_with_repository_name;
DROP VIEW IF EXISTS lsif_dumps_with_repository_name;
DROP VIEW IF EXISTS lsif_dumps;

ALTER TABLE lsif_uploads DROP COLUMN IF EXISTS indexer_version;

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

DROP VIEW IF EXISTS lsif_indexes_with_repository_name;
ALTER TABLE lsif_indexes DROP COLUMN IF EXISTS indexer_version;

CREATE VIEW lsif_indexes_with_repository_name AS
SELECT
    u.id,
    u.commit,
    u.queued_at,
    u.state,
    u.failure_message,
    u.started_at,
    u.finished_at,
    u.repository_id,
    u.process_after,
    u.num_resets,
    u.num_failures,
    u.docker_steps,
    u.root,
    u.indexer,
    u.indexer_args,
    u.outfile,
    u.log_contents,
    u.execution_logs,
    u.local_steps,
    r.name AS repository_name
FROM lsif_indexes u
JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;
