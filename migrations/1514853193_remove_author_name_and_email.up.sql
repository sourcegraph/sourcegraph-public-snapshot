BEGIN;
ALTER TABLE comments DROP COLUMN author_name;
ALTER TABLE comments DROP COLUMN author_email;
COMMIT;
