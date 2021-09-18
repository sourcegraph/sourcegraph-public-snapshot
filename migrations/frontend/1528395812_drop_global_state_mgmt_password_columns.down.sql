BEGIN;

ALTER TABLE global_state
ADD COLUMN IF NOT EXISTS mgmt_password_plaintext TEXT NOT NULL DEFAULT '',
ADD COLUMN IF NOT EXISTS mgmt_password_bcrypt TEXT NOT NULL DEFAULT '';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
