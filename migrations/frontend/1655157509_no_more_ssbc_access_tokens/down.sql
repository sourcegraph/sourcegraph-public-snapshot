ALTER TABLE batch_spec_workspace_execution_jobs ADD COLUMN IF NOT EXISTS access_token_id bigint REFERENCES access_tokens(id);

ALTER TABLE batch_spec_workspace_execution_jobs DROP CONSTRAINT batch_spec_workspace_execution_jobs_access_token_id_fkey;
ALTER TABLE batch_spec_workspace_execution_jobs
ADD CONSTRAINT batch_spec_workspace_execution_jobs_access_token_id_fkey
FOREIGN KEY (access_token_id)
REFERENCES access_tokens(id) ON DELETE SET NULL DEFERRABLE;

DROP VIEW batch_spec_workspace_execution_jobs_with_rank;
CREATE VIEW batch_spec_workspace_execution_jobs_with_rank AS SELECT j.id,
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
  j.access_token_id,
  j.queued_at,
  j.user_id,
  q.place_in_global_queue,
  q.place_in_user_queue
 FROM (batch_spec_workspace_execution_jobs j
 LEFT JOIN batch_spec_workspace_execution_queue q ON ((j.id = q.id)));
