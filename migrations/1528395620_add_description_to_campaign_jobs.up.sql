BEGIN;

ALTER TABLE campaign_jobs ADD COLUMN description text;

COMMIT;
