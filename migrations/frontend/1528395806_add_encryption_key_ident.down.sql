BEGIN;

ALTER TABLE user_external_accounts DROP COLUMN IF EXISTS encryption_key_id;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
