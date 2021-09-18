BEGIN;

ALTER TABLE IF EXISTS batch_spec_workspace_execution_jobs
  ADD COLUMN IF NOT EXISTS cancel boolean DEFAULT false NOT NULL;

CREATE INDEX IF NOT EXISTS batch_spec_workspace_execution_jobs_cancel
  ON batch_spec_workspace_execution_jobs (cancel);

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
