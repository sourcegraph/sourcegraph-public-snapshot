BEGIN;

ALTER TABLE IF EXISTS cm_trigger_jobs DROP COLUMN IF EXISTS execution_logs;
ALTER TABLE IF EXISTS cm_action_jobs  DROP COLUMN IF EXISTS execution_logs;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
