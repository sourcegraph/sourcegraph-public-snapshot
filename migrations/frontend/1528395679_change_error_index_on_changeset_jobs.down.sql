BEGIN;

DROP INDEX IF EXISTS changeset_jobs_error_not_null;

CREATE INDEX changeset_jobs_error ON changeset_jobs (error);

COMMIT;

