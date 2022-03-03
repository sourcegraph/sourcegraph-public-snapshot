ALTER TABLE insight_series
    DROP COLUMN IF EXISTS generation_method;
ALTER TABLE insight_series
    DROP COLUMN IF EXISTS just_in_time;
