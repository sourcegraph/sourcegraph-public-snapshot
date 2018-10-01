-- Remove the length constraint from the regexp so that it can be configurable.
ALTER TABLE orgs
	DROP CONSTRAINT org_name_valid_chars,
	ADD CONSTRAINT orgs_name_valid_chars CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$');
ALTER TABLE users
	DROP CONSTRAINT users_username_valid,
	ADD CONSTRAINT users_username_valid_chars CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|-(?=[a-zA-Z0-9]))*$');

-- Loosen the display name length constraint.
ALTER TABLE orgs
	DROP CONSTRAINT org_display_name_valid,
	ADD CONSTRAINT orgs_display_name_max_length CHECK (char_length(display_name) <= 255);
ALTER TABLE users
	DROP CONSTRAINT users_display_name_valid,
	ADD CONSTRAINT users_display_name_max_length CHECK (char_length(display_name) <= 255);

-- Add a length constraint to prevent extremely large inputs.
ALTER TABLE users
	ADD CONSTRAINT users_username_max_length CHECK (char_length(username) <= 255);
ALTER TABLE orgs
	ADD CONSTRAINT orgs_name_max_length CHECK (char_length(name) <= 255);
