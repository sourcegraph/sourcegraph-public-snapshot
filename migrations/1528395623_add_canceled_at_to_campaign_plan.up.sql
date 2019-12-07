BEGIN;

ALTER TABLE campaign_plans ADD COLUMN canceled_at timestamptz;

COMMIT;
