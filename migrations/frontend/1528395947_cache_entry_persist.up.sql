BEGIN;

ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS step_cache_results JSON NOT NULL DEFAULT '{}';

COMMIT;
