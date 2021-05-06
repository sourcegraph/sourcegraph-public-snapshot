BEGIN;

ALTER TABLE
    batch_changes_site_credentials
DROP COLUMN IF EXISTS
    encryption_key_id;

ALTER TABLE
    user_credentials
DROP COLUMN IF EXISTS
    encryption_key_id;

COMMIT;
