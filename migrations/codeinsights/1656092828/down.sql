-- Undo the changes made in the up migration
ALTER TABLE IF EXISTS series_points DROP COLUMN IF EXISTS path;
