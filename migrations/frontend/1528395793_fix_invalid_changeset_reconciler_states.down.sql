BEGIN;

UPDATE changesets SET reconciler_state = 'QUEUED' WHERE reconciler_state = 'queued';

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
