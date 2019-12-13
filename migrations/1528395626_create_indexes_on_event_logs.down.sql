BEGIN;

DROP INDEX IF EXISTS event_logs_source;
DROP INDEX IF EXISTS event_logs_date_trunc_timestamp;

COMMIT;
