-- +++
-- parent: 1000000002
-- +++

BEGIN;

DELETE FROM series_points; -- affects dev environments only, others never had data in this table.
ALTER TABLE series_points ALTER COLUMN series_id TYPE text;
ALTER TABLE series_points ALTER COLUMN series_id SET NOT NULL;

-- Give series_id a btree index since we'll be filtering on it very frequently.
CREATE INDEX series_points_series_id_btree ON series_points USING btree (series_id);

COMMIT;
