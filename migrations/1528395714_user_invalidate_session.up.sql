BEGIN;

-- Default to now, since we can't have null times
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS invalidated_sessions_at timestamp with time zone DEFAULT now() NOT NULL;
-- Update the invalidated sessions at to be when the user was created
UPDATE users SET invalidated_sessions_at = created_at;

COMMIT;
