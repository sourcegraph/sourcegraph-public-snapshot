BEGIN;

ALTER TABLE campaign_plans DROP COLUMN IF EXISTS arguments;
ALTER TABLE campaign_plans DROP COLUMN IF EXISTS canceled_at;
ALTER TABLE campaign_plans DROP COLUMN IF EXISTS campaign_type;

COMMIT;
