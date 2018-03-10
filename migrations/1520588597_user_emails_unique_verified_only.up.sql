-- Makes the user_emails uniqueness constraint looser. First we add the two looser constraints, then
-- we remove the stricter constraint.

-- Prevent multiple users from having the same verified email address. PostgreSQL doesn't support
-- partial unique indexes, but this has the same effect.
ALTER TABLE user_emails ADD CONSTRAINT user_emails_unique_verified_email EXCLUDE (email WITH =) WHERE (verified_at IS NOT NULL);

-- Prevent a user from having duplicates of the same email address associated with their own
-- account.
ALTER TABLE user_emails ADD CONSTRAINT user_emails_no_duplicates_per_user UNIQUE (user_id, email);

-- Remove the stricter constraint.
ALTER TABLE user_emails DROP CONSTRAINT user_emails_email_key;
