BEGIN;

CREATE INDEX IF NOT EXISTS event_logs_source ON event_logs(source);
CREATE INDEX IF NOT EXISTS event_logs_timestamp_at_UTC ON event_logs(DATE(timestamp AT TIME ZONE 'UTC'));

COMMIT;
