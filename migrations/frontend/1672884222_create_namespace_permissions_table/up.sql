CREATE TABLE IF NOT EXISTS namespace_permissions (
    id SERIAL PRIMARY KEY,
    namespace text NOT NULL,
    resource_id integer NOT NULL,
    action text NOT NULL,

    user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,

    created_at timestamp with time zone DEFAULT now() NOT NULL,

    CONSTRAINT namespace_not_blank CHECK (namespace <> ''::text),
    CONSTRAINT action_not_blank CHECK (action <> ''::text),

    CONSTRAINT unique_resource_permission UNIQUE (namespace, resource_id, action, user_id)
);