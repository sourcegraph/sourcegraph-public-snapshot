BEGIN;

ALTER TABLE campaign_jobs ADD COLUMN base_ref text;
UPDATE campaign_jobs SET base_ref = 'master';
ALTER TABLE campaign_jobs ALTER COLUMN base_ref SET NOT NULL;
ALTER TABLE campaign_jobs ADD CONSTRAINT campaign_jobs_base_ref_check CHECK (base_ref != '');

COMMIT;
