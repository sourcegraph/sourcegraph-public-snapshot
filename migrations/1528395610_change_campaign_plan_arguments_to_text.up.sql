BEGIN;

-- Since it's not easy to convert `jsonb` to `text`, we simply drop the
-- column since the table has been unused so far.
ALTER TABLE campaign_plans DROP COLUMN arguments;
ALTER TABLE campaign_plans
  ADD COLUMN arguments text NOT NULL;

COMMIT;
