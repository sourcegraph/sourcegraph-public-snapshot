BEGIN;

UPDATE insight_series
SET sample_interval_unit = 'MONTH'
WHERE sample_interval_unit IS NULL;

UPDATE insight_series
SET sample_interval_value = 1
WHERE sample_interval_value IS NULL;

ALTER TABLE insight_series
    ALTER COLUMN sample_interval_unit SET DEFAULT 'MONTH',
    ALTER COLUMN sample_interval_unit SET NOT NULL,
    ALTER COLUMN sample_interval_value SET DEFAULT '1',
    ALTER COLUMN sample_interval_value SET NOT NULL;

COMMIT;
