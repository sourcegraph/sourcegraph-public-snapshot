
BEGIN;

ALTER TABLE event_logs
DROP COLUMN cohort_id;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
