BEGIN;

ALTER TABLE campaign_plans DROP COLUMN arguments;
ALTER TABLE campaign_plans
  ADD COLUMN arguments jsonb NOT NULL DEFAULT '{}'
  CHECK (jsonb_typeof(arguments) = 'object');

COMMIT;
