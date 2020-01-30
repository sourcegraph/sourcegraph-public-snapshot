BEGIN;

ALTER TABLE user_emails DROP COLUMN IF EXISTS last_verification_sent_at;

COMMIT;
