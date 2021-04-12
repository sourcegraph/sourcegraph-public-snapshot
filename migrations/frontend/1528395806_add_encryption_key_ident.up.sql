BEGIN;

ALTER TABLE user_external_accounts ADD COLUMN IF NOT EXISTS encryption_key_id text NOT NULL DEFAULT '';

COMMIT;
