BEGIN;

ALTER TABLE campaign_jobs DROP COLUMN base_ref;

COMMIT;
