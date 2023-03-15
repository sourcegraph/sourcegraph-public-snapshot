DROP VIEW batch_spec_workspace_execution_jobs_with_rank;
CREATE VIEW batch_spec_workspace_execution_jobs_with_rank AS (
    SELECT
        j.id,
        j.batch_spec_workspace_id,
        j.state,
        j.failure_message,
        j.started_at,
        j.finished_at,
        j.process_after,
        j.num_resets,
        j.num_failures,
        j.execution_logs,
        j.worker_hostname,
        j.last_heartbeat_at,
        j.created_at,
        j.updated_at,
        j.cancel,
        j.queued_at,
        j.user_id,
        q.place_in_global_queue,
        q.place_in_user_queue
    FROM
        batch_spec_workspace_execution_jobs j
    LEFT JOIN batch_spec_workspace_execution_queue q ON j.id = q.id
);

ALTER TABLE batch_spec_workspace_execution_jobs DROP COLUMN IF EXISTS version;
