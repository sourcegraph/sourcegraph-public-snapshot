BEGIN;

ALTER TABLE batch_spec_workspaces
  ADD COLUMN IF NOT EXISTS cached_result_found BOOLEAN NOT NULL DEFAULT FALSE;

UPDATE
  batch_spec_workspaces
SET
  cached_result_found = true
WHERE
  batch_spec_execution_cache_entry_id IS NOT NULL;

ALTER TABLE batch_spec_workspaces
  DROP COLUMN IF EXISTS batch_spec_execution_cache_entry_id;

DELETE FROM
  batch_spec_execution_cache_entries e1
WHERE
  EXISTS (SELECT 1 FROM batch_spec_execution_cache_entries e2 WHERE e1.key = e2.key AND e1.id != e2.id);

ALTER TABLE batch_spec_execution_cache_entries
  ADD CONSTRAINT batch_spec_execution_cache_entries_key_unique UNIQUE (key);

COMMIT;
