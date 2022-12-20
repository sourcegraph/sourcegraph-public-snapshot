-- read only roles that come with every sourcegraph instance
INSERT INTO roles VALUES (1, 'DEFAULT', NOW(), NULL, TRUE);
INSERT INTO roles VALUES (2, 'SITE_ADMINISTRATOR', NOW(), NULL, TRUE);

WITH default_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = 'DEFAULT'
),
site_admin_role AS MATERIALIZED (
    SELECT id FROM roles WHERE name = 'SITE_ADMINISTRATOR'
),
users_with_roles AS (
    -- this assigns the DEFAULT role to every user in the database
	SELECT id AS user_id, (SELECT id FROM default_role) AS role_id, NOW() AS created_at FROM users
        UNION
    -- this assigns the SITE_ADMINISTRATOR role to every user in the database
    SELECT id AS user_id, (SELECT id FROM site_admin_role) AS role_id, NOW() AS created_at FROM users WHERE users.site_admin
)
INSERT INTO user_roles (SELECT * FROM users_with_roles);
