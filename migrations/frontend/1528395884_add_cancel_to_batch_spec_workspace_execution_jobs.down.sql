BEGIN;

DROP INDEX IF EXISTS batch_spec_workspace_execution_jobs_cancel;

ALTER TABLE IF EXISTS batch_spec_workspace_execution_jobs
  DROP COLUMN IF EXISTS cancel;

COMMIT;
