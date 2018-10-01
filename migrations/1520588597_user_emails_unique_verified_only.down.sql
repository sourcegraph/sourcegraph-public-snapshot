-- If we truly rolled back to the stricter constraint, some rows might not validate. We never
-- depended on the stricter validation, so just "roll back" a constraint with the old name
-- but the same more lenient definition.
ALTER TABLE user_emails RENAME CONSTRAINT user_emails_unique_verified_email TO user_emails_email_key;

ALTER TABLE user_emails DROP CONSTRAINT user_emails_no_duplicates_per_user;
