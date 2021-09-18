BEGIN;

-- PR #22080 landed a number of backwards-incompatible API docs data changes and given how early
-- stages API docs is, we don't care to maintain backwards compat with the old data and choose to
-- instead start from scratch with indexing again (not many repos have been indexed with API docs,
-- anyway.)
TRUNCATE lsif_data_documentation_pages;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE codeintel_schema_migrations SET dirty = 'f'
COMMIT;
