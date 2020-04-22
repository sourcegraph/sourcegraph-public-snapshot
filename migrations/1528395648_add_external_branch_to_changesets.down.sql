BEGIN;

ALTER TABLE changesets
DROP COLUMN IF EXISTS external_branch;

COMMIT;
