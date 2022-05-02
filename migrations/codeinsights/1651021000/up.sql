CREATE TYPE series_sort_mode_enum AS ENUM (
    'RESULT_COUNT',
    'LEXICOGRAPHICAL',
    'DATE_ADDED'
);

CREATE TYPE series_sort_direction_enum AS ENUM (
    'ASC',
    'DESC'
);

ALTER TABLE IF EXISTS insight_view
    ADD COLUMN IF NOT EXISTS series_sort_mode series_sort_mode_enum,
    ADD COLUMN IF NOT EXISTS series_sort_direction series_sort_direction_enum,
    ADD COLUMN IF NOT EXISTS series_limit INT;
