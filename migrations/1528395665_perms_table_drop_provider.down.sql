BEGIN;

ALTER TABLE user_permissions ADD COLUMN provider TEXT;
ALTER TABLE repo_permissions ADD COLUMN provider TEXT;

COMMIT;
