BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions DROP CONSTRAINT batch_spec_executions_has_1_namespace;
ALTER TABLE IF EXISTS batch_spec_executions DROP COLUMN IF EXISTS namespace_user_id;
ALTER TABLE IF EXISTS batch_spec_executions DROP COLUMN IF EXISTS namespace_org_id;

-- Clear the dirty flag in case the operator timed out and isn't around to clear it.
UPDATE schema_migrations SET dirty = 'f'
COMMIT;
