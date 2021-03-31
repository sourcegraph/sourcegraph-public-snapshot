BEGIN;

ALTER TABLE user_external_accounts DROP COLUMN IF EXISTS encryption_key_id;

COMMIT;
