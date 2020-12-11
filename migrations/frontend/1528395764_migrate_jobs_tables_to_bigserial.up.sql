BEGIN;

ALTER TABLE cm_trigger_jobs
    ALTER COLUMN id SET DATA TYPE bigint;
ALTER SEQUENCE cm_trigger_jobs_id_seq as bigint;

ALTER TABLE cm_action_jobs
    ALTER COLUMN id SET DATA TYPE bigint,
    ALTER COLUMN trigger_event SET DATA type int8;
ALTER SEQUENCE cm_action_jobs_id_seq as bigint;

COMMIT;
