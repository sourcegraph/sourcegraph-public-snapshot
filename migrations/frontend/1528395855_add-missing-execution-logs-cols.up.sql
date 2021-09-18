BEGIN;

ALTER TABLE IF EXISTS cm_trigger_jobs ADD COLUMN IF NOT EXISTS execution_logs JSON[];
ALTER TABLE IF EXISTS cm_action_jobs  ADD COLUMN IF NOT EXISTS execution_logs JSON[];

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
