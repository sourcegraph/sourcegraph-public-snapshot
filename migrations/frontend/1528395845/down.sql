BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions DROP CONSTRAINT batch_spec_executions_has_1_namespace;
ALTER TABLE IF EXISTS batch_spec_executions DROP COLUMN IF EXISTS namespace_user_id;
ALTER TABLE IF EXISTS batch_spec_executions DROP COLUMN IF EXISTS namespace_org_id;

COMMIT;
