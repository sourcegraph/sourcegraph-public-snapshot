CREATE TABLE IF NOT EXISTS user_roles (
    user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    role_id integer REFERENCES roles(id) ON DELETE CASCADE DEFERRABLE,

    created_at timestamp with time zone DEFAULT now() NOT NULL,

    PRIMARY KEY (user_id, role_id)
);
