ALTER TABLE users
	DROP CONSTRAINT users_username_valid_chars,
	ADD CONSTRAINT users_username_valid CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$');
ALTER TABLE orgs
	DROP CONSTRAINT orgs_name_valid_chars,
	ADD CONSTRAINT org_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9])){0,38}$');

ALTER TABLE orgs
	DROP CONSTRAINT orgs_display_name_max_length,
	ADD CONSTRAINT org_display_name_valid CHECK (char_length(display_name) <= 64);
ALTER TABLE users
	DROP CONSTRAINT users_display_name_max_length,
	ADD CONSTRAINT users_display_name_valid CHECK (char_length(display_name) <= 64);

ALTER TABLE users DROP CONSTRAINT users_username_max_length;
ALTER TABLE orgs DROP CONSTRAINT orgs_name_max_length;
