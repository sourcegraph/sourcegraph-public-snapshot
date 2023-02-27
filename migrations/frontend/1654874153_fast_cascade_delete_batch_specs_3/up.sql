-- Create index for foreign key reverse lookups. This is required for cascading deletes.
CREATE INDEX CONCURRENTLY IF NOT EXISTS batch_spec_workspace_execution_jobs_batch_spec_workspace_id ON batch_spec_workspace_execution_jobs (batch_spec_workspace_id);
