BEGIN;

ALTER TABLE campaign_plans DROP COLUMN canceled_at;

COMMIT;
