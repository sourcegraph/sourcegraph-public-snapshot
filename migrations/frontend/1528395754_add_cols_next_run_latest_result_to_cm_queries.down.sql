BEGIN;
ALTER TABLE cm_queries
    DROP COLUMN IF EXISTS next_run,
    DROP COLUMN IF EXISTS latest_result;
COMMIT;
