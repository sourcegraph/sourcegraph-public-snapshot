BEGIN;

ALTER TABLE user_permissions DROP COLUMN full_synced_at;
ALTER TABLE repo_permissions DROP COLUMN full_synced_at;

COMMIT;
