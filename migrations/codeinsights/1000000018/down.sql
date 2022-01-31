BEGIN;

ALTER TABLE IF EXISTS insight_series
    ALTER COLUMN sample_interval_unit DROP DEFAULT,
    ALTER COLUMN sample_interval_unit DROP NOT NULL,
    ALTER COLUMN sample_interval_value DROP DEFAULT,
    ALTER COLUMN sample_interval_value DROP NOT NULL;

COMMIT;
