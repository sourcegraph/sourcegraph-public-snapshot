BEGIN;

ALTER TABLE IF EXISTS batch_spec_workspace_execution_jobs
  ADD COLUMN IF NOT EXISTS cancel boolean DEFAULT false NOT NULL;

COMMIT;
