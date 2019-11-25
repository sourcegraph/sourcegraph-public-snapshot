BEGIN;

ALTER TABLE changeset_jobs DROP CONSTRAINT changeset_jobs_campaign_id_fkey,
  ADD CONSTRAINT changeset_jobs_campaign_id_fkey
    FOREIGN KEY (campaign_id)
    REFERENCES campaigns(id)
    DEFERRABLE INITIALLY IMMEDIATE;

COMMIT;
