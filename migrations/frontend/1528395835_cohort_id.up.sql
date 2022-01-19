-- +++
-- parent: 1528395834
-- +++


BEGIN;

ALTER TABLE event_logs
ADD COLUMN cohort_id date;

COMMIT;
