BEGIN;

ALTER TABLE IF EXISTS batch_spec_workspace_execution_jobs
  DROP COLUMN IF EXISTS access_token_id;

ALTER TABLE IF EXISTS access_tokens
  DROP COLUMN IF EXISTS internal;

COMMIT;
