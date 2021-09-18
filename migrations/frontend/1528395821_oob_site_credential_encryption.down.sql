BEGIN;

-- We need to leave the new column in place here for the OOB down migration, so
-- no changes here.

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
