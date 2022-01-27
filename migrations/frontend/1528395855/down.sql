BEGIN;

ALTER TABLE IF EXISTS cm_trigger_jobs DROP COLUMN IF EXISTS execution_logs;
ALTER TABLE IF EXISTS cm_action_jobs  DROP COLUMN IF EXISTS execution_logs;

COMMIT;
