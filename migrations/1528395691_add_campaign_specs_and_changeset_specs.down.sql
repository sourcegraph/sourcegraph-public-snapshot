BEGIN;

ALTER TABLE IF EXISTS campaigns DROP COLUMN IF EXISTS campaign_spec_id;

DROP TABLE IF EXISTS changeset_specs;
DROP TABLE IF EXISTS campaign_specs;

COMMIT;
