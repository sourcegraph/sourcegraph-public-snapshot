-- +++
-- parent: 1528395854
-- +++

BEGIN;

ALTER TABLE IF EXISTS cm_trigger_jobs ADD COLUMN IF NOT EXISTS execution_logs JSON[];
ALTER TABLE IF EXISTS cm_action_jobs  ADD COLUMN IF NOT EXISTS execution_logs JSON[];

COMMIT;
