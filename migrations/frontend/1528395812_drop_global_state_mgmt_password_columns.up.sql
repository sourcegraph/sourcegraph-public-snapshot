BEGIN;

ALTER TABLE global_state
DROP COLUMN IF EXISTS mgmt_password_plaintext,
DROP COLUMN IF EXISTS mgmt_password_bcrypt;

COMMIT;
