BEGIN;

ALTER TABLE insight_series
    DROP COLUMN generation_method;
ALTER TABLE insight_series
    DROP COLUMN just_in_time;

COMMIT;
