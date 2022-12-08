CREATE TABLE IF NOT EXISTS role_permissions (
    role_id integer REFERENCES roles(id) ON DELETE CASCADE DEFERRABLE,
    permission_id integer REFERENCES permissions(id) ON DELETE CASCADE DEFERRABLE,

    PRIMARY KEY (permission_id, role_id)
);
