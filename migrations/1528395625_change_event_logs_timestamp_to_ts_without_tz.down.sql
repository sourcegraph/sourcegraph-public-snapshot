BEGIN;

ALTER TABLE event_logs
ALTER COLUMN timestamp TYPE timestamp with time zone;

COMMIT;
