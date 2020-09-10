BEGIN;

UPDATE campaign_jobs SET description = '' WHERE description IS NULL;
ALTER TABLE campaign_jobs ALTER COLUMN description SET NOT NULL;

COMMIT;
