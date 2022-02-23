ALTER TABLE IF EXISTS insight_series
    ADD COLUMN IF NOT EXISTS generation_method TEXT,
    ADD COLUMN IF NOT EXISTS just_in_time      BOOL;

COMMENT ON COLUMN insight_series.generation_method is 'Specifies the execution method for how this series is generated. This helps the system understand how to generate the time series data.';
COMMENT ON COLUMN insight_series.just_in_time is 'Specifies if the series should be resolved just in time at query time, or recorded in background processing.';

-- This is just formalizing some of the logic that exists. It seems a little brittle, but that's why we are doing this now :)
UPDATE insight_series
SET generation_method =
        CASE
            WHEN (sample_interval_unit = 'MONTH' AND sample_interval_value = 0) THEN 'language-stats'
            WHEN (generated_from_capture_groups IS TRUE) THEN 'search-compute'
            ELSE 'search'
            END;

ALTER TABLE IF EXISTS insight_series
    ALTER COLUMN generation_method SET NOT NULL;

-- This is using some slightly funky logic to ensure we are setting the just_in_time flag appropriately based on any existing series.
UPDATE insight_series
SET just_in_time =
        CASE
            WHEN (generation_method = 'search' AND (CARDINALITY(repositories) = 0 OR repositories IS NULL)) THEN FALSE
            WHEN (generation_method = 'search' AND (CARDINALITY(repositories) > 0)) THEN TRUE
            WHEN (generation_method = 'language-stats' AND (CARDINALITY(repositories) > 0)) THEN TRUE
            WHEN (generation_method = 'search-compute' AND (CARDINALITY(repositories) = 0 OR repositories IS NULL))
                THEN FALSE
            WHEN (generation_method = 'search-compute' AND (CARDINALITY(repositories) > 0)) THEN TRUE
            ELSE FALSE
            END;

ALTER TABLE IF EXISTS insight_series
    ALTER COLUMN just_in_time SET NOT NULL,
    ALTER COLUMN just_in_time SET DEFAULT FALSE;
