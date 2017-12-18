BEGIN;
ALTER TABLE users ADD COLUMN passwd text;
ALTER TABLE users ADD COLUMN email_code text;
COMMIT;
