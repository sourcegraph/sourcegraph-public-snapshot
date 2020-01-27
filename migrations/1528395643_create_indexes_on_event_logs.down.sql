BEGIN;

DROP INDEX IF EXISTS event_logs_source;
DROP INDEX IF EXISTS event_logs_timestamp_at_UTC;

COMMIT;
