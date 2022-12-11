CREATE TABLE IF NOT EXISTS roles (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,

    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW() NOT NULL,
    deleted_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT name_not_blank CHECK ((name <> ''::text)),
    CONSTRAINT roles_name UNIQUE (name)
);

COMMENT ON COLUMN roles.name IS 'The uniquely identifying name of the role.';
