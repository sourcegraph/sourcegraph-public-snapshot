BEGIN;

ALTER TABLE changesets
DROP COLUMN IF EXISTS external_state;

ALTER TABLE changesets
DROP COLUMN IF EXISTS external_review_state;

ALTER TABLE changesets
DROP COLUMN IF EXISTS external_check_state;

COMMIT;
