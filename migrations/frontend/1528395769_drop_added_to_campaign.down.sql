BEGIN;

ALTER TABLE changesets ADD COLUMN IF NOT EXISTS added_to_campaign boolean NOT NULL DEFAULT false;

COMMIT;
