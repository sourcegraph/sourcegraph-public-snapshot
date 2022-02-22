ALTER TABLE IF EXISTS series_points
    DROP COLUMN IF EXISTS capture;

ALTER TABLE IF EXISTS series_points_snapshots
    DROP COLUMN IF EXISTS capture;
