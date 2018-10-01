
ALTER TABLE orgs
	DROP CONSTRAINT org_name_valid_chars,
	ADD CONSTRAINT org_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$');

ALTER TABLE users
	DROP CONSTRAINT users_username_valid,
	ADD CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$');

COMMIT;