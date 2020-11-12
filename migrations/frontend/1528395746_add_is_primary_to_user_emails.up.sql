BEGIN;

ALTER TABLE user_emails ADD COLUMN is_primary bool DEFAULT false NOT NULL;

-- Use our old logic to set the initial primary address.
-- From this point we expect it to be set from the UI.

UPDATE user_emails SET is_primary = true
WHERE (user_id, email) IN
(SELECT DISTINCT ON (user_id) user_id, email FROM user_emails ORDER BY user_id, (verified_at IS NOT NULL) DESC, created_at ASC, email ASC);

-- A user can only have one primary address
CREATE UNIQUE INDEX user_emails_user_id_is_primary_idx ON user_emails (user_id, is_primary)
WHERE is_primary = true;

COMMIT;
