ALTER TABLE settings RENAME COLUMN author_auth_id TO author_user_id;

ALTER TABLE settings RENAME COLUMN author_user_id TO author_user_id_old;
ALTER TABLE settings ADD COLUMN author_user_id integer REFERENCES users (id) ON DELETE RESTRICT;
UPDATE settings SET author_user_id=(SELECT users.id FROM users WHERE users.auth_id=settings.author_user_id_old);
CREATE TABLE settings_bkup_1514702776 AS (SELECT * FROM settings WHERE author_user_id IS NULL);
DELETE FROM settings WHERE author_user_id IS NULL;
ALTER TABLE settings ALTER COLUMN author_user_id SET NOT NULL;
ALTER TABLE settings DROP COLUMN author_user_id_old;
