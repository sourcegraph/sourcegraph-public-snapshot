BEGIN;

ALTER TABLE IF EXISTS event_logs DROP COLUMN IF EXISTS public_argument;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
