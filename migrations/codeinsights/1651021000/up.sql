ALTER TABLE IF EXISTS insight_view
    ADD COLUMN IF NOT EXISTS sort_series_by TEXT default 'HIGHEST_RESULT_COUNT'::text,
    ADD COLUMN IF NOT EXISTS display_num_series INT default 20;
