BEGIN;
ALTER TABLE cm_trigger_jobs
    DROP COLUMN IF EXISTS query_string,
    DROP COLUMN IF EXISTS results;
COMMIT;
