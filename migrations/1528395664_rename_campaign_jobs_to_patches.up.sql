BEGIN;

ALTER TABLE campaign_jobs RENAME TO patches;

ALTER TABLE changeset_jobs RENAME COLUMN campaign_job_id TO patch_id;

COMMIT;
