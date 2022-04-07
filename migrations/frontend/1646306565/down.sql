ALTER TABLE repo_pending_permissions ADD COLUMN IF NOT EXISTS user_ids bytea NOT NULL DEFAULT '\x'::bytea;
ALTER TABLE user_pending_permissions ADD COLUMN IF NOT EXISTS object_ids bytea NOT NULL DEFAULT '\x'::bytea;
