BEGIN;

ALTER TABLE batch_spec_workspaces
  ADD COLUMN IF NOT EXISTS batch_spec_execution_cache_entry_id INTEGER REFERENCES batch_spec_execution_cache_entries(id) DEFERRABLE;

COMMIT;
