BEGIN;

-- Perform a table swap - create a new table and rename the hypertable
CREATE TABLE series_points_vanilla (LIKE series_points INCLUDING ALL);
ALTER TABLE series_points RENAME TO series_points_timescale;
ALTER TABLE series_points_vanilla RENAME TO series_points;

-- Copy all of the data and insert into the new table.
INSERT INTO series_points (SELECT * FROM series_points_timescale);

-- Drop the old hypertable (the extension will propagate and drop all of the hypertable stuff)
DROP TABLE series_points_timescale CASCADE;

-- Last, remove the extension
DROP EXTENSION IF EXISTS timescaledb;

COMMIT;
