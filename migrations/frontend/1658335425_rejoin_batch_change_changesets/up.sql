-- Set up the new table and its dependent type.

DROP TYPE IF EXISTS batch_change_changesets_archived;

CREATE TYPE batch_change_changesets_archived AS ENUM (
  'pending',
  'archived'
);

CREATE TABLE IF NOT EXISTS batch_change_changesets (
  batch_change_id BIGINT NOT NULL,
  changeset_id BIGINT NOT NULL,
  detach BOOLEAN NOT NULL DEFAULT FALSE,
  archived batch_change_changesets_archived NULL DEFAULT NULL,
  PRIMARY KEY (batch_change_id, changeset_id)
);

CREATE INDEX IF NOT EXISTS
  batch_change_changesets_detach_idx
ON
  batch_change_changesets (archived);

-- Migrate data from changesets.batch_change_ids into the new table.

INSERT INTO
  batch_change_changesets
  (batch_change_id, changeset_id, detach, archived)
SELECT
  (assoc.key)::BIGINT AS batch_change_id,
  changesets.id AS changeset_id,
  COALESCE((assoc.value->'detach')::BOOLEAN, FALSE) AS detach,
  CASE
    WHEN (assoc.value->'isArchived')::BOOLEAN = TRUE THEN 'archived'::batch_change_changesets_archived
    WHEN (assoc.value->'archive')::BOOLEAN = TRUE THEN 'pending'::batch_change_changesets_archived
  END AS archived
FROM 
  changesets,
  JSONB_EACH(changesets.batch_change_ids) AS assoc;

-- Recreate the reconciler_changesets view.
--
-- TODO: Note that this omits the batch_change_ids column; the relevant
-- scanChangeset method has to be updated anyway.

DROP VIEW IF EXISTS reconciler_changesets;

CREATE VIEW reconciler_changesets AS
SELECT c.id,
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
INNER JOIN batch_change_changesets bcc ON bcc.changeset_id = c.id
WHERE r.deleted_at IS NULL AND EXISTS (
    SELECT 1
    FROM batch_changes
    LEFT JOIN users namespace_user ON batch_changes.namespace_user_id = namespace_user.id
    LEFT JOIN orgs namespace_org ON batch_changes.namespace_org_id = namespace_org.id
    WHERE namespace_user.deleted_at IS NULL AND namespace_org.deleted_at IS NULL
);

-- Drop the changesets.batch_change_ids column.

ALTER TABLE changesets DROP COLUMN batch_change_ids;
