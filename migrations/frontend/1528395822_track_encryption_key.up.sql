BEGIN;

ALTER TABLE
    batch_changes_site_credentials
ADD COLUMN IF NOT EXISTS
    encryption_key_id TEXT NOT NULL DEFAULT '';

ALTER TABLE
    user_credentials
ADD COLUMN IF NOT EXISTS
    encryption_key_id TEXT NOT NULL DEFAULT '';

COMMIT;
