BEGIN;

DROP INDEX IF EXISTS changeset_jobs_error;

CREATE INDEX changeset_jobs_error_not_null ON changeset_jobs ((error IS NOT NULL));

COMMIT;

