BEGIN;

ALTER TABLE batch_spec_workspaces
  ADD COLUMN IF NOT EXISTS cached_result_found BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE batch_spec_workspaces
SET cached_result_found = true
WHERE batch_spec_execution_cache_entry_id IS NOT NULL;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS batch_spec_execution_cache_entry_id;

COMMIT;
