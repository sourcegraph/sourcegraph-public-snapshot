ALTER TABLE users DROP CONSTRAINT check_external_id;
UPDATE users SET external_id=CONCAT('native:', COALESCE((SELECT email FROM user_emails WHERE user_emails.user_id=users.id ORDER BY created_at ASC, email ASC LIMIT 1), CONCAT(username, '@example.com')));
UPDATE users SET external_provider='native' WHERE external_id LIKE 'native:%';
UPDATE users SET external_provider='auth0' WHERE external_id LIKE 'auth0|%';
