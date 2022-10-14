CREATE TABLE IF NOT EXISTS executor_secrets (
    id SERIAL PRIMARY KEY,
    key text NOT NULL,
    value bytea NOT NULL,
    scope text NOT NULL,
    encryption_key_id text,
    namespace_user_id integer REFERENCES users(id) ON DELETE CASCADE,
    namespace_org_id integer REFERENCES orgs(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT NOW(),
    updated_at timestamp with time zone NOT NULL DEFAULT NOW(),
    creator_id integer REFERENCES users(id) ON DELETE SET NULL
);

COMMENT ON COLUMN executor_secrets.creator_id IS 'NULL, if the user has been deleted.';

CREATE UNIQUE INDEX executor_secrets_unique_key_namespace ON executor_secrets (key, namespace_user_id, namespace_org_id, scope);
CREATE UNIQUE INDEX executor_secrets_unique_key_global ON executor_secrets(key, scope) WHERE namespace_user_id IS NULL AND namespace_org_id IS NULL;
