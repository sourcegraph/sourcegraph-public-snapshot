BEGIN;

ALTER TABLE changeset_jobs DROP CONSTRAINT changeset_jobs_campaign_job_id_fkey,
  ADD CONSTRAINT changeset_jobs_campaign_job_id_fkey
    FOREIGN KEY (campaign_job_id)
    REFERENCES campaign_jobs(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

ALTER TABLE changeset_jobs DROP CONSTRAINT changeset_jobs_changeset_id_fkey,
  ADD CONSTRAINT changeset_jobs_changeset_id_fkey
    FOREIGN KEY (changeset_id)
    REFERENCES changesets(id)
    ON DELETE CASCADE DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
