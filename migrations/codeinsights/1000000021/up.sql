BEGIN;

ALTER TABLE insight_dirty_queries
    DROP CONSTRAINT insight_dirty_queries_insight_series_id_fkey,
    ADD CONSTRAINT insight_dirty_queries_insight_series_id_fkey FOREIGN KEY (insight_series_id) REFERENCES insight_series (id) ON DELETE CASCADE;

COMMIT;
