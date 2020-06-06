BEGIN;

ALTER TABLE changesets DROP COLUMN IF EXISTS created_by_campaign;
ALTER TABLE changesets DROP COLUMN IF EXISTS added_to_campaign;

COMMIT;
