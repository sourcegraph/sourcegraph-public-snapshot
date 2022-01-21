BEGIN;

ALTER TABLE cm_trigger_jobs
    DROP COLUMN result_payload;

COMMIT;
