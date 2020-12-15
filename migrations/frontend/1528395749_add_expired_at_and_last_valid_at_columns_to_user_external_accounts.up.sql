BEGIN;

ALTER TABLE user_external_accounts ADD COLUMN IF NOT EXISTS expired_at TIMESTAMPTZ;
ALTER TABLE user_external_accounts ADD COLUMN IF NOT EXISTS last_valid_at TIMESTAMPTZ;

COMMIT;
