BEGIN;

ALTER TABLE campaign_plans ADD COLUMN arguments text NOT NULL;
ALTER TABLE campaign_plans ADD COLUMN campaign_type text NOT NULL;
ALTER TABLE campaign_plans ADD COLUMN canceled_at timestamp with time zone;

COMMIT;
