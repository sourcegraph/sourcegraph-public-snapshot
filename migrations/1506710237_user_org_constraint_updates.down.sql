ALTER TABLE org_members
	DROP CONSTRAINT org_members_references_orgs;

ALTER TABLE orgs
	DROP CONSTRAINT org_name_valid_chars,
	ADD CONSTRAINT org_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9_\-]+$');

ALTER TABLE users
	DROP CONSTRAINT users_display_name_valid,
	DROP CONSTRAINT users_username_valid,
	ADD CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,37}[a-zA-Z0-9])?$');
