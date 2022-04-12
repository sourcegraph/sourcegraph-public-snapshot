ALTER TABLE repo_pending_permissions DROP COLUMN IF EXISTS user_ids;
ALTER TABLE user_pending_permissions DROP COLUMN IF EXISTS object_ids;
