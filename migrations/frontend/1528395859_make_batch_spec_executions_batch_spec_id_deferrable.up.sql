
BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ALTER CONSTRAINT batch_spec_executions_batch_spec_id_fkey DEFERRABLE;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
