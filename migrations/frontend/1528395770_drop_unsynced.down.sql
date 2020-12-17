BEGIN;

ALTER TABLE changesets ADD COLUMN IF NOT EXISTS unsynced boolean NOT NULL DEFAULT false;
UPDATE changesets SET unsynced = true, publication_state = 'PUBLISHED' WHERE publication_state = 'UNPUBLISHED' AND current_spec_id IS NULL AND owned_by_campaign_id IS NULL;

COMMIT;
