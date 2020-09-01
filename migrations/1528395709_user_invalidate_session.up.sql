BEGIN;

-- Create this as null, since we don't expect the user to ever have invalidated their sessionsinvalid
ALTER TABLE IF EXISTS users ADD COLUMN IF NOT EXISTS invalidated_sessions_at timestamp with time zone;

COMMIT;
