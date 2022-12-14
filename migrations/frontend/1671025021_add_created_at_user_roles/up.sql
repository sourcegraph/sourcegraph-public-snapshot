ALTER TABLE user_roles
    ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

UPDATE user_roles SET created_at = NOW();

ALTER TABLE user_roles
    ALTER COLUMN created_at SET NOT NULL;
