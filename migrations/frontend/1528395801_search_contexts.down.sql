BEGIN;

DROP TABLE IF EXISTS search_context_repos;

DROP TABLE IF EXISTS search_contexts;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
