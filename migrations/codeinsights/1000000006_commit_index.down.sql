BEGIN;

DROP TABLE IF EXISTS commit_index;
DROP TABLE IF EXISTS commit_index_metadata;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeinsights_schema_migrations SET dirty = 'f'
COMMIT;
