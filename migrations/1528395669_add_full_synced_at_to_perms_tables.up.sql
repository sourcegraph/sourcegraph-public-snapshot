BEGIN;

ALTER TABLE user_permissions ADD COLUMN full_synced_at TIMESTAMPTZ;
ALTER TABLE repo_permissions ADD COLUMN full_synced_at TIMESTAMPTZ;

COMMIT;
