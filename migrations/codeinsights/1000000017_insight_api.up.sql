BEGIN;

CREATE TYPE time_unit AS ENUM ('HOUR', 'DAY', 'WEEK', 'MONTH', 'YEAR');
ALTER TABLE insight_series
    DROP COLUMN IF EXISTS recording_interval_days,
    ADD COLUMN repositories TEXT[],
    ADD COLUMN sample_interval_unit time_unit,
    ADD COLUMN sample_interval_value int
;

ALTER TABLE insight_view
    ADD COLUMN default_filter_include_repo_regex text,
    ADD COLUMN default_filter_exclude_repo_regex text
;

