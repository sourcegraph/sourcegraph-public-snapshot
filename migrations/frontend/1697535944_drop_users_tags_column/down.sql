ALTER TABLE users ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}'::TEXT[];
