BEGIN;

ALTER TABLE patches ADD COLUMN IF NOT EXISTS diff_stat_added integer;
ALTER TABLE patches ADD COLUMN IF NOT EXISTS diff_stat_changed integer;
ALTER TABLE patches ADD COLUMN IF NOT EXISTS diff_stat_deleted integer;

COMMIT;
