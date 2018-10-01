CREATE TABLE user_emails(
	user_id integer NOT NULL REFERENCES users (id),
	email citext NOT NULL UNIQUE,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	verification_code text,
	verified_at timestamp with time zone
);
INSERT INTO user_emails(user_id, email, created_at, verification_code, verified_at)
	SELECT id AS user_id, email, created_at, email_code AS verification_code,
		(CASE WHEN email_code IS NULL THEN now() ELSE null END) AS verified_at
	FROM users;
ALTER TABLE users DROP COLUMN email;
ALTER TABLE users DROP COLUMN email_code;
