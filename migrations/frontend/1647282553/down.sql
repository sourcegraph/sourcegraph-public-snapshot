--
-- Drop views that added new column

DROP VIEW IF EXISTS external_service_sync_jobs_with_next_sync_at;
DROP VIEW IF EXISTS lsif_dumps_with_repository_name;
DROP VIEW IF EXISTS lsif_uploads_with_repository_name;
DROP VIEW IF EXISTS lsif_dumps;
DROP VIEW IF EXISTS reconciler_changesets;

--
-- Recreate old views

CREATE VIEW external_service_sync_jobs_with_next_sync_at AS
SELECT j.id,
    j.state,
    j.failure_message,
    j.started_at,
    j.finished_at,
    j.process_after,
    j.num_resets,
    j.num_failures,
    j.execution_logs,
    j.external_service_id,
    e.next_sync_at
FROM external_services e
JOIN external_service_sync_jobs j ON e.id = j.external_service_id;

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
FROM lsif_dumps u JOIN repo r ON r.id = u.repository_id
WHERE r.deleted_at IS NULL;

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

CREATE VIEW reconciler_changesets AS
SELECT c.id,
    c.batch_change_ids,
    c.repo_id,
    c.created_at,
    c.updated_at,
    c.metadata,
    c.external_id,
    c.external_service_type,
    c.external_deleted_at,
    c.external_branch,
    c.external_updated_at,
    c.external_state,
    c.external_review_state,
    c.external_check_state,
    c.diff_stat_added,
    c.diff_stat_changed,
    c.diff_stat_deleted,
    c.sync_state,
    c.current_spec_id,
    c.previous_spec_id,
    c.publication_state,
    c.owned_by_batch_change_id,
    c.reconciler_state,
    c.failure_message,
    c.started_at,
    c.finished_at,
    c.process_after,
    c.num_resets,
    c.closing,
    c.num_failures,
    c.log_contents,
    c.execution_logs,
    c.syncer_error,
    c.external_title,
    c.worker_hostname,
    c.ui_publication_state,
    c.last_heartbeat_at,
    c.external_fork_namespace
FROM changesets c
JOIN repo r ON r.id = c.repo_id
WHERE r.deleted_at IS NULL AND EXISTS (
    SELECT 1
    FROM batch_changes
    LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id
    LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id
    WHERE c.batch_change_ids ? batch_changes.id::text AND namespace_user.deleted_at IS NULL AND namespace_org.deleted_at IS NULL
);

--
-- Drop new columns

ALTER TABLE batch_spec_resolution_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE batch_spec_workspace_execution_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE changeset_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE changesets DROP COLUMN IF EXISTS queued_at;
ALTER TABLE cm_action_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE cm_trigger_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE external_service_sync_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE insights_query_runner_jobs DROP COLUMN IF EXISTS queued_at;
ALTER TABLE lsif_uploads DROP COLUMN IF EXISTS queued_at;
