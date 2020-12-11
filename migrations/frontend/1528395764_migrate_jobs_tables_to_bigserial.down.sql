BEGIN;

ALTER TABLE cm_trigger_jobs
    ALTER COLUMN id SET DATA TYPE int;
ALTER SEQUENCE cm_trigger_jobs_id_seq as int;

ALTER TABLE cm_action_jobs
    ALTER COLUMN id SET DATA TYPE int,
    ALTER COLUMN trigger_event SET DATA type int4;
ALTER SEQUENCE cm_action_jobs_id_seq as int;

COMMIT;
