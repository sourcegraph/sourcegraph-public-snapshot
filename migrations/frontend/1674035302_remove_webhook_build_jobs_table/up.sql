DROP INDEX IF EXISTS webhook_build_jobs_queued_at_idx;
DROP INDEX IF EXISTS webhook_build_jobs_state;
DROP TABLE IF EXISTS webhook_build_jobs;
DROP SEQUENCE IF EXISTS webhook_build_jobs_id_seq;
