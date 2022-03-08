ALTER TABLE IF EXISTS series_points
    ADD COLUMN IF NOT EXISTS capture TEXT;

ALTER TABLE IF EXISTS series_points_snapshots
    ADD COLUMN IF NOT EXISTS capture TEXT;
