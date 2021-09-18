BEGIN;

ALTER TABLE external_services ADD COLUMN IF NOT EXISTS encryption_key_id text NOT NULL DEFAULT '';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
