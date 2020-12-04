BEGIN;
ALTER TABLE cm_action_jobs
    ADD COLUMN IF NOT EXISTS trigger_event int,
    ADD CONSTRAINT cm_action_jobs_trigger_event_fk
        FOREIGN KEY (trigger_event)
            REFERENCES cm_trigger_jobs (id)
            ON DELETE CASCADE;

ALTER TABLE cm_trigger_jobs
    ADD COLUMN IF NOT EXISTS num_results int;
COMMIT;
