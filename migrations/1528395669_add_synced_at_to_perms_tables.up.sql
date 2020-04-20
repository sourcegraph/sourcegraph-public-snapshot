BEGIN;

ALTER TABLE user_permissions ADD COLUMN synced_at TIMESTAMPTZ;
ALTER TABLE repo_permissions ADD COLUMN synced_at TIMESTAMPTZ;

COMMIT;
