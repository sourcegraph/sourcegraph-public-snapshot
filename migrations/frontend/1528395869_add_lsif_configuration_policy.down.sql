BEGIN;

DROP TABLE IF EXISTS lsif_configuration_policies;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
