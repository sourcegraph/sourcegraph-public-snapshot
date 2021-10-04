BEGIN;

ALTER TABLE IF EXISTS batch_spec_workspace_execution_jobs
  ADD COLUMN IF NOT EXISTS access_token_id bigint REFERENCES access_tokens(id) ON DELETE CASCADE DEFERRABLE DEFAULT NULL;

ALTER TABLE IF EXISTS access_tokens
  ADD COLUMN IF NOT EXISTS internal boolean DEFAULT FALSE;

COMMIT;
