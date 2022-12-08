CREATE TABLE IF NOT EXISTS user_roles (
    user_id integer REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    role_id integer REFERENCES roles(id) ON DELETE CASCADE DEFERRABLE,

    PRIMARY KEY (user_id, role_id)
);
