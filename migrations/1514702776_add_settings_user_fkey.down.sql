ALTER TABLE settings RENAME COLUMN author_user_id TO author_user_id_new;
ALTER TABLE settings ADD COLUMN author_user_id text;
UPDATE settings SET author_user_id=(SELECT users.auth_id FROM users WHERE users.id=settings.author_user_id_new);
ALTER TABLE settings ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE settings DROP COLUMN author_user_id_new;
INSERT INTO settings(id, org_id, author_user_id, contents, created_at, user_id)
	SELECT id, org_id, author_user_id_old, contents, created_at, user_id FROM settings_bkup_1514702776;
DROP TABLE settings_bkup_1514702776;

ALTER TABLE settings RENAME COLUMN author_user_id TO author_auth_id;
