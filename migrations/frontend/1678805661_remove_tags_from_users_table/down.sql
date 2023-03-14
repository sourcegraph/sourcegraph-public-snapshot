ALTER TABLE IF EXISTS users
    ADD COLUMN IF NOT EXISTS tags text[] DEFAULT '{}'::text[];
