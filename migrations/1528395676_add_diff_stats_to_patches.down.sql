BEGIN;

ALTER TABLE patches DROP COLUMN IF EXISTS diff_stat_added;
ALTER TABLE patches DROP COLUMN IF EXISTS diff_stat_changed;
ALTER TABLE patches DROP COLUMN IF EXISTS diff_stat_deleted;

COMMIT;
