ALTER TABLE org_members RENAME COLUMN user_id TO user_id_old;
ALTER TABLE org_members ADD COLUMN user_id integer REFERENCES users (id) ON DELETE RESTRICT;
UPDATE org_members SET user_id=(SELECT users.id FROM users WHERE users.auth_id=org_members.user_id_old);
CREATE TABLE org_members_bkup_1514536731 AS (SELECT * FROM org_members WHERE user_id IS NULL);
DELETE FROM org_members WHERE user_id IS NULL;
ALTER TABLE org_members ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE org_members DROP COLUMN user_id_old;
ALTER TABLE org_members ADD UNIQUE (org_id, user_id);
