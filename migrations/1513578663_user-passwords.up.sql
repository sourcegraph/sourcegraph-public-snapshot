ALTER TABLE users ADD COLUMN passwd text;
ALTER TABLE users ADD COLUMN email_code text;
ALTER TABLE users ADD COLUMN passwd_reset_code text;
ALTER TABLE users ADD COLUMN passwd_reset_time timestamp with time zone;
