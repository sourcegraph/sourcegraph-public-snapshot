
BEGIN;

ALTER TABLE event_logs
DROP COLUMN cohort_id;

COMMIT;
