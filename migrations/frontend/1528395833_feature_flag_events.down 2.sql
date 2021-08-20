BEGIN;

ALTER TABLE event_logs
DROP COLUMN feature_flags;

COMMIT;
