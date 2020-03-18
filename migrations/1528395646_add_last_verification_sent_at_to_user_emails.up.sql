BEGIN;

ALTER TABLE user_emails ADD COLUMN last_verification_sent_at timestamp with time zone;

COMMIT;
