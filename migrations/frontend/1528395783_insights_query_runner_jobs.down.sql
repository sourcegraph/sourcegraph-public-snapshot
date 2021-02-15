BEGIN;

DROP INDEX IF EXISTS insights_query_runner_jobs_state_btree;
DROP TABLE IF EXISTS insights_query_runner_jobs;

COMMIT;
