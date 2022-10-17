CREATE INDEX CONCURRENTLY IF NOT EXISTS batch_spec_workspace_execution_jobs_last_dequeue ON batch_spec_workspace_execution_jobs (user_id, started_at DESC);
