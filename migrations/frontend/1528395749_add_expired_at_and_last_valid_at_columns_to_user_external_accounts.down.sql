BEGIN;

ALTER TABLE user_external_accounts DROP COLUMN IF EXISTS expired_at;
ALTER TABLE user_external_accounts DROP COLUMN IF EXISTS last_valid_at;

COMMIT;
