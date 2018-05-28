ALTER TABLE org_members RENAME COLUMN user_id TO user_id_new;
ALTER TABLE org_members ADD COLUMN user_id text;
UPDATE org_members SET user_id=(SELECT users.auth_id FROM users WHERE users.id=org_members.user_id_new);
ALTER TABLE org_members ALTER COLUMN user_id SET NOT NULL;
ALTER TABLE org_members DROP COLUMN user_id_new;
ALTER TABLE org_members ADD UNIQUE (org_id, user_id);
INSERT INTO org_members(org_id, user_id, created_at, updated_at) SELECT org_id, user_id_old, created_at, updated_at FROM org_members_bkup_1514536731;
DROP TABLE org_members_bkup_1514536731;
