BEGIN;

DROP TABLE IF EXISTS insight_view_series;
DROP TABLE IF EXISTS insight_view;
DROP TABLE IF EXISTS insight_series;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeinsights_schema_migrations SET dirty = 'f'
COMMIT;
