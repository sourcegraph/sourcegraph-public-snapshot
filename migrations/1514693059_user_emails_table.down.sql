ALTER TABLE users ADD COLUMN email citext UNIQUE;
ALTER TABLE users ADD COLUMN email_code text;
UPDATE users SET email=(SELECT email FROM user_emails WHERE user_emails.user_id=users.id ORDER BY created_at ASC, email ASC LIMIT 1);
UPDATE users SET email_code=(SELECT verification_code FROM user_emails WHERE user_emails.user_id=users.id ORDER BY created_at ASC, email ASC LIMIT 1);
DROP TABLE user_emails;
UPDATE users SET email=CONCAT('no-email-', id, '@example.com') WHERE email IS NULL; -- account for future users w/o email
ALTER TABLE users ALTER COLUMN email SET NOT NULL;
