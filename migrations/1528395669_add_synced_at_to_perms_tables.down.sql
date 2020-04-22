BEGIN;

ALTER TABLE user_permissions DROP COLUMN synced_at;
ALTER TABLE repo_permissions DROP COLUMN synced_at;

COMMIT;
