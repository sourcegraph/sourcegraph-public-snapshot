ALTER TABLE org_members
	ADD CONSTRAINT org_members_references_orgs
	FOREIGN KEY (org_id)
	REFERENCES orgs (id) ON DELETE RESTRICT;

ALTER TABLE orgs
	DROP CONSTRAINT org_name_valid_chars,
	ADD CONSTRAINT org_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,36}[a-zA-Z0-9])?$');

ALTER TABLE users
	DROP CONSTRAINT users_username_valid,
	ADD CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9]([a-zA-Z0-9\-]{0,36}[a-zA-Z0-9])?$'),
	ADD CONSTRAINT users_display_name_valid CHECK (char_length(display_name) <= 64);
