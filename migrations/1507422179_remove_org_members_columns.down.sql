
ALTER TABLE org_members ADD COLUMN "email" TEXT;
ALTER TABLE org_members ADD COLUMN "username" TEXT;
ALTER TABLE org_members ADD COLUMN "display_name" TEXT;
ALTER TABLE org_members ADD COLUMN "avatar_url" TEXT;

UPDATE org_members
SET email = users.email,
    username = users.username,
    display_name = users.display_name,
    avatar_url = users.avatar_url
FROM (
    SELECT auth0_id, email, username, display_name, avatar_url FROM users
) AS users
WHERE org_members.user_id = users.auth0_id;

ALTER TABLE org_members ALTER COLUMN "email" SET NOT NULL;
ALTER TABLE org_members ALTER COLUMN "username" SET NOT NULL;
ALTER TABLE org_members ALTER COLUMN "display_name" SET NOT NULL;

ALTER TABLE org_members
	ADD CONSTRAINT org_members_org_id_user_name_key UNIQUE (org_id, username),
	ADD CONSTRAINT org_members_org_id_user_email_key UNIQUE (org_id, email);

