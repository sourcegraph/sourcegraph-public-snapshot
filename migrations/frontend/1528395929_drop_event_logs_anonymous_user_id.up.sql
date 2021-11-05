BEGIN;

-- This index is not needed and about 2GB in size.
DROP INDEX IF EXISTS event_logs_anonymous_user_id;

COMMIT;
