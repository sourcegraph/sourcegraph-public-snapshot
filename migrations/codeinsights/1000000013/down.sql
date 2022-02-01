BEGIN;

DROP TABLE IF EXISTS insight_view_grants;

ALTER TABLE insight_view_series
    DROP CONSTRAINT IF EXISTS insight_view_series_insight_view_id_fkey;

ALTER TABLE insight_view_series
    ADD CONSTRAINT insight_view_series_insight_view_id_fkey
        FOREIGN KEY (insight_view_id) REFERENCES insight_view;

COMMIT;
