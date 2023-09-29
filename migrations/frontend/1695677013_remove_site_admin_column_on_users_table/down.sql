ALTER TABLE users
	ADD COLUMN IF NOT EXISTS site_admin boolean DEFAULT false NOT NULL;

UPDATE users
    SET site_admin = TRUE
    WHERE id IN (
        SELECT ur.user_id
        FROM user_roles ur
        JOIN roles r ON ur.role_id = r.id
        WHERE r.name = 'SITE_ADMINISTRATOR'
    );
