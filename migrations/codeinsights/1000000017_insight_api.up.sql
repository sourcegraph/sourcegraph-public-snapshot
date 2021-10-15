BEGIN;

CREATE TYPE time_unit AS ENUM ('HOUR', 'DAY', 'WEEK', 'MONTH', 'YEAR');
ALTER TABLE insight_series
    DROP COLUMN IF EXISTS recording_interval_days,
    ADD COLUMN repositories TEXT[],
    ADD COLUMN sample_interval_unit time_unit,
    add column sample_interval_value int
;

alter table insight_view
    add column default_filter_include_repo_regex text,
    add column default_filter_exclude_repo_regex text
;

