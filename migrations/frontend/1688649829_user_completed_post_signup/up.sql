ALTER TABLE users ADD COLUMN IF NOT EXISTS completed_post_signup BOOLEAN NOT NULL DEFAULT FALSE;

-- We mark all users that have verified their email as having completed_post_signup
-- the post-signup flow
UPDATE users
SET completed_post_signup = TRUE
FROM user_emails
WHERE
    user_emails.user_id = users.id
AND
    user_emails.verified_at IS NOT NULL;
