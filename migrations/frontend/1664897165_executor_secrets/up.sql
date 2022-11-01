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

-- Enforce uniqueness of the key in a given namespace.
CREATE UNIQUE INDEX IF NOT EXISTS executor_secrets_unique_key_namespace_user ON executor_secrets (key, namespace_user_id, scope) WHERE namespace_user_id IS NOT NULL;
CREATE UNIQUE INDEX IF NOT EXISTS executor_secrets_unique_key_namespace_org ON executor_secrets (key, namespace_org_id, scope) WHERE namespace_org_id IS NOT NULL;
-- Enforce uniqueness of the key in the global namespace. NULL is a fun type :)
CREATE UNIQUE INDEX IF NOT EXISTS executor_secrets_unique_key_global ON executor_secrets(key, scope) WHERE namespace_user_id IS NULL AND namespace_org_id IS NULL;

CREATE TABLE IF NOT EXISTS executor_secret_access_logs (
    id SERIAL PRIMARY KEY,
    executor_secret_id integer NOT NULL REFERENCES executor_secrets(id) ON DELETE CASCADE,
    user_id integer NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at timestamp with time zone NOT NULL DEFAULT NOW()
);
