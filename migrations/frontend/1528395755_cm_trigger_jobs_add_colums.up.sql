BEGIN;
ALTER TABLE cm_trigger_jobs
    ADD COLUMN IF NOT EXISTS query_string text,
    ADD COLUMN IF NOT EXISTS results boolean;
COMMIT;
