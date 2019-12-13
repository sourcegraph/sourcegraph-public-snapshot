BEGIN;

CREATE INDEX IF NOT EXISTS event_logs_source ON event_logs(source);
CREATE INDEX IF NOT EXISTS event_logs_date_trunc_timestamp ON event_logs(DATE_TRUNC('day', timestamp));

COMMIT;
