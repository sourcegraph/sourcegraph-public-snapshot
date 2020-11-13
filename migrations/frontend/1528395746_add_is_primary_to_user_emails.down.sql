BEGIN;

DROP INDEX IF EXISTS user_emails_user_id_is_primary_idx;
ALTER TABLE user_emails DROP COLUMN IF EXISTS is_primary;

COMMIT;
