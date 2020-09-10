BEGIN;

ALTER TABLE patches RENAME TO campaign_jobs;

ALTER TABLE changeset_jobs RENAME COLUMN patch_id TO campaign_job_id;

COMMIT;
