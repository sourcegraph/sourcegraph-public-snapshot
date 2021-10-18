BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

ALTER TABLE insight_series
    DROP COLUMN IF EXISTS repositories,
    DROP COLUMN IF EXISTS sample_interval_unit,
    DROP COLUMN IF EXISTS sample_interval_value,
    ADD COLUMN IF NOT EXISTS recording_interval_days int;

ALTER TABLE insight_view
    DROP COLUMN IF EXISTS default_filter_include_repo_regex,
    DROP COLUMN IF EXISTS default_filter_exclude_repo_regex;
COMMIT;
