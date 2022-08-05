DO $$
BEGIN
    CREATE TYPE series_sort_mode_enum AS ENUM (
        'RESULT_COUNT',
        'LEXICOGRAPHICAL',
        'DATE_ADDED'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

DO $$
BEGIN
    CREATE TYPE series_sort_direction_enum AS ENUM (
        'ASC',
        'DESC'
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

ALTER TABLE IF EXISTS insight_view
    ADD COLUMN IF NOT EXISTS series_sort_mode series_sort_mode_enum,
    ADD COLUMN IF NOT EXISTS series_sort_direction series_sort_direction_enum,
    ADD COLUMN IF NOT EXISTS series_limit INT;
