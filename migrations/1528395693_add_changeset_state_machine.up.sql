BEGIN;

-- First we need to add something we forgot in the last migration: the foreign
-- key that allows us to wire up changesets with changeset_specs.
ALTER TABLE changesets
  ADD COLUMN IF NOT EXISTS changeset_spec_id bigint REFERENCES changeset_specs(id) DEFERRABLE;

-- Now we add the 'state' field to changesets.
-- See ./internal/campaigns/types.go for the possible values here:
--   - UNPUBLISHED
--   - PUBLISHING
--   - ERRORED
--   - SYNCED
-- We use UNPUBLISHED was the default value
ALTER TABLE changesets ADD COLUMN IF NOT EXISTS state text DEFAULT 'UNPUBLISHED';

-- Since changesets can now be created in an "unpublished" state, we need to
-- make the following columns nullable:
ALTER TABLE changesets ALTER COLUMN external_id DROP NOT NULL;
ALTER TABLE changesets ALTER COLUMN metadata DROP NOT NULL;
ALTER TABLE changesets ALTER COLUMN external_service_type DROP NOT NULL;

-- But before switching to the new flow every changeset we had has been created
-- on the code host and is synced.
UPDATE changesets SET state = 'SYNCED';

-- Now we delete all changeset_jobs so that we have nothing in a
-- temporary state lying around.
DELETE FROM changeset_jobs;

-- And we unset the patch_set_id from every campaign so they look like "manual"
-- campaigns and the patch sets will get cleaned up because they're expired.
UPDATE campaigns SET patch_set_id = NULL WHERE patch_set_id IS NOT NULL;


COMMIT;
