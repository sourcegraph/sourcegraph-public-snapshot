BEGIN;

DROP INDEX IF EXISTS changeset_jobs_state_idx;
DROP INDEX IF EXISTS changeset_jobs_group_idx;

COMMIT;
