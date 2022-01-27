BEGIN;

ALTER TABLE IF EXISTS batch_spec_executions ADD COLUMN IF NOT EXISTS namespace_user_id integer REFERENCES users(id) DEFERRABLE;
ALTER TABLE IF EXISTS batch_spec_executions ADD COLUMN IF NOT EXISTS namespace_org_id integer REFERENCES orgs(id) DEFERRABLE;
UPDATE batch_spec_executions SET namespace_user_id = user_id;
ALTER TABLE IF EXISTS batch_spec_executions ADD CONSTRAINT batch_spec_executions_has_1_namespace CHECK ((namespace_user_id IS NULL) <> (namespace_org_id IS NULL));

COMMIT;
