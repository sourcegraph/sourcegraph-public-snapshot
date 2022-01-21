-- +++
-- parent: 1528395968
-- +++

BEGIN;

ALTER TABLE cm_trigger_jobs
    ADD COLUMN IF NOT EXISTS result_payload JSONB;

COMMIT;
