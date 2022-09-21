ALTER TABLE IF EXISTS series_points DROP COLUMN IF EXISTS metadata_id;
ALTER TABLE IF EXISTS series_points_snapshots DROP COLUMN IF EXISTS metadata_id;
DROP TABLE IF EXISTS metadata;