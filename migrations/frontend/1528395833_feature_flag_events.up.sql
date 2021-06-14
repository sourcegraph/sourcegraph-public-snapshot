BEGIN;

ALTER TABLE event_logs
ADD COLUMN feature_flags jsonb;

COMMIT;
