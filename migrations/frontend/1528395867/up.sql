-- +++
-- parent: 1528395866
-- +++

BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ADD COLUMN IF NOT EXISTS cancel BOOL DEFAULT FALSE;

COMMIT;
