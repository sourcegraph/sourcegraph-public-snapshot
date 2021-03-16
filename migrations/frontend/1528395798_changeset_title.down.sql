BEGIN;

DROP INDEX IF EXISTS changesets_external_title_idx;
ALTER TABLE changesets DROP COLUMN IF EXISTS external_title;

COMMIT;
