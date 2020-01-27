BEGIN;

ALTER TABLE event_logs ALTER COLUMN argument TYPE text;

COMMIT;
