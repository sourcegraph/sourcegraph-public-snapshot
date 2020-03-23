BEGIN;

-- Remove the provider column.
ALTER TABLE user_permissions DROP COLUMN provider;
ALTER TABLE repo_permissions DROP COLUMN provider;

COMMIT;
