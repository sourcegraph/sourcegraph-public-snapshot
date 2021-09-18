BEGIN;

ALTER TABLE IF EXISTS event_logs ADD COLUMN IF NOT EXISTS public_argument JSONB DEFAULT '{}'::jsonb NOT NULL;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
