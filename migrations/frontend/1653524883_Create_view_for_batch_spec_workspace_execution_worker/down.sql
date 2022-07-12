DROP VIEW IF EXISTS batch_spec_workspace_execution_queue;

ALTER TABLE batch_spec_workspace_execution_jobs DROP COLUMN IF EXISTS user_id;

DROP INDEX IF EXISTS batch_spec_workspace_execution_jobs_user_id;

DROP INDEX IF EXISTS batch_spec_workspace_execution_jobs_state;
