BEGIN;

-- Note that we have to regenerate the reconciler_changesets view, as the SELECT
-- statement in the view definition isn't refreshed when the fields change within the
-- changesets table.
DROP VIEW IF EXISTS
    reconciler_changesets;

ALTER TABLE changesets
    ADD COLUMN IF NOT EXISTS commit_verification jsonb DEFAULT '{}'::jsonb NOT NULL;

CREATE VIEW reconciler_changesets AS
SELECT c.id,
    c.batch_change_ids,
    c.repo_id,
    c.queued_at,
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
    c.commit_verification,
    c.diff_stat_added,
    c.diff_stat_deleted,
    c.sync_state,
    c.current_spec_id,
    c.previous_spec_id,
    c.publication_state,
    c.owned_by_batch_change_id,
    c.reconciler_state,
    c.computed_state,
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
    c.external_fork_name,
    c.external_fork_namespace,
    c.detached_at,
    c.previous_failure_message
FROM changesets c
JOIN repo r ON r.id = c.repo_id
WHERE r.deleted_at IS NULL AND EXISTS (
    SELECT 1
    FROM batch_changes
        LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id
        LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id
    WHERE c.batch_change_ids ? batch_changes.id::text AND namespace_user.deleted_at IS NULL AND namespace_org.deleted_at IS NULL
    );

COMMIT;
