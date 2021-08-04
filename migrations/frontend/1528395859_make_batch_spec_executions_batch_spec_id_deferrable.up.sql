
BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ALTER CONSTRAINT batch_spec_executions_batch_spec_id_fkey DEFERRABLE;

COMMIT;
