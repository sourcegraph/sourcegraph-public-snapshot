BEGIN;

ALTER TABLE global_state
DROP COLUMN IF EXISTS mgmt_password_plaintext,
DROP COLUMN IF EXISTS mgmt_password_bcrypt;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
