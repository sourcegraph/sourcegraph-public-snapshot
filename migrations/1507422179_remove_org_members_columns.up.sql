
INSERT INTO users (auth0_id, email, username, display_name, avatar_url) (
    SELECT org_members.user_id, org_members.email, org_members.username, org_members.display_name, org_members.avatar_url FROM org_members LEFT OUTER JOIN users ON org_members.user_id = users.auth0_id WHERE users.auth0_id IS NULL
);

ALTER TABLE org_members
	DROP CONSTRAINT org_members_org_id_user_name_key,
	DROP CONSTRAINT org_members_org_id_user_email_key;

ALTER TABLE org_members DROP COLUMN "email";
ALTER TABLE org_members DROP COLUMN "username";
ALTER TABLE org_members DROP COLUMN "display_name";
ALTER TABLE org_members DROP COLUMN "avatar_url";

