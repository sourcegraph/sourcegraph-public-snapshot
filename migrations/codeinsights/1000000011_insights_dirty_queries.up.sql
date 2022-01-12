-- +++
-- parent: 1000000010
-- +++


BEGIN;

-- Insert migration here. See README.md. Highlights:
--  * Always use IF EXISTS. eg: DROP TABLE IF EXISTS global_dep_private;
--  * All migrations must be backward-compatible. Old versions of Sourcegraph
--    need to be able to read/write post migration.
--  * Historically we advised against transactions since we thought the
--    migrate library handled it. However, it does not! /facepalm

CREATE TABLE insight_dirty_queries
(
    id                SERIAL NOT NULL,
    insight_series_id INT,
    query             TEXT NOT NULL,
    dirty_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    reason            TEXT NOT NULL,
    for_time          TIMESTAMP NOT NULL,
    PRIMARY KEY (id),
    FOREIGN KEY (insight_series_id) REFERENCES insight_series(id)
);

CREATE INDEX insight_dirty_queries_insight_series_id_fk_idx ON insight_dirty_queries (insight_series_id);

COMMENT ON TABLE insight_dirty_queries IS 'Stores queries that were unsuccessful or otherwise flagged as incomplete or incorrect.';

COMMENT ON COLUMN insight_dirty_queries.query IS 'Sourcegraph query string that was executed.';
COMMENT ON COLUMN insight_dirty_queries.dirty_at IS 'Timestamp when this query was marked dirty.';
COMMENT ON COLUMN insight_dirty_queries.reason IS 'Human readable string indicating the reason the query was marked dirty.';
COMMENT ON COLUMN insight_dirty_queries.for_time IS 'Timestamp for which the original data point was recorded or intended to be recorded.';

COMMIT;
