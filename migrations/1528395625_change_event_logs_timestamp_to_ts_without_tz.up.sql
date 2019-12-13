BEGIN;

ALTER TABLE event_logs
ALTER COLUMN timestamp TYPE timestamp without time zone;

COMMIT;
