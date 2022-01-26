BEGIN;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS batch_spec_execution_cache_entry_id;

COMMIT;
