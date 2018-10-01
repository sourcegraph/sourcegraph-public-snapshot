ALTER TABLE orgs
	ADD CONSTRAINT org_name_unique UNIQUE (name),
	ADD CONSTRAINT org_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9_\-]+$');

ALTER TABLE org_members
	ADD UNIQUE (org_id, user_id),
	ADD UNIQUE (org_id, user_name),
	ADD UNIQUE (org_id, user_email);
