BEGIN;

ALTER TABLE campaigns ADD COLUMN thread_ids jsonb
NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(thread_ids) = 'object');

CREATE INDEX campaigns_thread_ids_gin_idx ON campaigns
USING GIN (thread_ids);

ALTER TABLE threads DROP COLUMN campaign_id;
ALTER TABLE threads ADD COLUMN campaign_ids jsonb
NOT NULL DEFAULT '{}' CHECK (jsonb_typeof(campaign_ids) = 'object');

CREATE INDEX threads_campaign_ids_gin_idx ON threads
USING GIN (campaign_ids);

COMMIT;
