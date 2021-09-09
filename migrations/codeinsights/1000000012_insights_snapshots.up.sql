
BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE series_points_snapshots
(
    LIKE series_points INCLUDING DEFAULTS INCLUDING CONSTRAINTS INCLUDING INDEXES
);

COMMENT ON TABLE series_points_snapshots is 'Stores ephemeral snapshot data of insight recordings.';

alter table insight_series
    add last_snapshot_at timestamp default (CURRENT_TIMESTAMP - '10 years'::interval);

alter table insight_series
    add next_snapshot_after timestamp default CURRENT_TIMESTAMP;

COMMIT;
