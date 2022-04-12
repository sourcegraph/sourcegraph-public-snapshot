CREATE UNIQUE INDEX IF NOT EXISTS batch_changes_unique_user_id ON batch_changes (name, namespace_user_id) WHERE namespace_user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS batch_changes_unique_org_id ON batch_changes (name, namespace_org_id) WHERE namespace_org_id IS NOT NULL;
