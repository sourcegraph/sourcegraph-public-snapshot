BEGIN;

CREATE INDEX IF NOT EXISTS event_logs_anonymous_user_id ON event_logs (anonymous_user_id);

COMMIT;
