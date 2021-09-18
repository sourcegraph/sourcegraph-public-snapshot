BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ADD COLUMN IF NOT EXISTS cancel BOOL DEFAULT FALSE;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
