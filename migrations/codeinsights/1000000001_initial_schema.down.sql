BEGIN;

DROP TABLE IF EXISTS series_points;
DROP TABLE IF EXISTS repo_names;
DROP TABLE IF EXISTS metadata;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeinsights_schema_migrations SET dirty = 'f'
COMMIT;
