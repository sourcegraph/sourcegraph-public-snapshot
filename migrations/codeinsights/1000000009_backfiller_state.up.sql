-- +++
-- parent: 1000000008
-- +++


BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

ALTER TABLE insight_series ADD COLUMN backfill_queued_at TIMESTAMP;
COMMENT ON COLUMN insight_series.series_id IS
    'Timestamp that this series completed a full repository iteration for backfill. This flag has limited semantic value, and only means it tried to queue up queries for each repository. It does not guarantee success on those queries.';

COMMIT;
