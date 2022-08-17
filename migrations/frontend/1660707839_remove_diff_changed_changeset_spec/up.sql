-- We update the `diff_stat_added` and `diff_stat_deleted` to reflect the way git calculate diffs.
-- When calculating diffs, we only care about the added & deleted lines
UPDATE changeset_specs
    SET diff_stat_added = diff_stat_added + diff_stat_changed,
    diff_stat_deleted = diff_stat_deleted + diff_stat_changed
    WHERE diff_stat_changed != 0;

ALTER TABLE IF EXISTS changeset_specs DROP COLUMN IF EXISTS diff_stat_changed;
