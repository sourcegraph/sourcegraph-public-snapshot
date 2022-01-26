-- +++
-- parent: 1528395963
-- +++

BEGIN;

ALTER TABLE batch_spec_workspaces ADD COLUMN IF NOT EXISTS step_cache_results JSONB NOT NULL DEFAULT '{}';

COMMIT;
