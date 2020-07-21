BEGIN;

ALTER TABLE changesets DROP COLUMN IF EXISTS changeset_spec_id;
ALTER TABLE changesets DROP COLUMN IF EXISTS state;
ALTER TABLE changesets ALTER COLUMN external_id SET NOT NULL;
ALTER TABLE changesets ALTER COLUMN metadata SET NOT NULL;
ALTER TABLE changesets ALTER COLUMN external_service_type SET NOT NULL;

COMMIT;
