
BEGIN;

ALTER TABLE event_logs
ADD COLUMN cohort_id date;

COMMIT;
