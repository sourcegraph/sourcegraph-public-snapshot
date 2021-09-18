BEGIN;

DROP TRIGGER versions_insert ON versions;
DROP FUNCTION versions_insert_row_trigger;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
