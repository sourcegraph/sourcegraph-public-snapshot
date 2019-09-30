BEGIN;

ALTER TABLE orgs
    DROP CONSTRAINT orgs_name_valid_chars,
    ADD CONSTRAINT orgs_name_valid_chars
        CHECK (name ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*$');

ALTER TABLE users
    DROP CONSTRAINT users_username_valid_chars,
    ADD CONSTRAINT users_username_valid_chars
        CHECK (username ~ '^[a-zA-Z0-9](?:[a-zA-Z0-9]|[-.](?=[a-zA-Z0-9]))*$');

COMMIT;
