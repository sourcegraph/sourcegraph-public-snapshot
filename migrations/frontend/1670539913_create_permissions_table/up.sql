CREATE TABLE IF NOT EXISTS permissions (
    id SERIAL PRIMARY KEY,
    namespace text NOT NULL,
    action TEXT NOT NULL,

    created_at timestamp with time zone DEFAULT now() NOT NULL,

    CONSTRAINT namespace_not_blank CHECK (namespace <> ''::text),
    CONSTRAINT action_not_blank CHECK (action <> ''::text)
);

-- Enforce uniqueness of actions in a given namespace
CREATE UNIQUE INDEX IF NOT EXISTS permissions_unique_namespace_action ON permissions (namespace, action);
