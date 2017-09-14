ALTER TABLE orgs
	DROP CONSTRAINT org_name_unique,
	DROP CONSTRAINT org_name_valid_chars;

ALTER TABLE org_members
	DROP CONSTRAINT org_members_org_id_user_id_key,
	DROP CONSTRAINT org_members_org_id_user_name_key,
	DROP CONSTRAINT org_members_org_id_user_email_key;
