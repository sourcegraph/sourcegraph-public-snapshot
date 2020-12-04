BEGIN;
ALTER TABLE cm_action_jobs
    DROP COLUMN IF EXISTS trigger_event;
ALTER TABLE cm_trigger_jobs
    DROP COLUMN IF EXISTS num_results;
COMMIT;
