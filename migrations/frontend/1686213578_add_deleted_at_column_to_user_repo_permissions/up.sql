ALTER TABLE user_repo_permissions
    ADD COLUMN IF NOT EXISTS deleted_at timestamp with time zone DEFAULT NULL;
