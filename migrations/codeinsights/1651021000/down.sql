ALTER TABLE IF EXISTS insight_view
    DROP COLUMN IF EXISTS series_sort_mode,
    DROP COLUMN IF EXISTS series_sort_direction,
    DROP COLUMN IF EXISTS series_limit;

DROP TYPE series_sort_mode_enum;
DROP TYPE series_sort_direction_enum;
