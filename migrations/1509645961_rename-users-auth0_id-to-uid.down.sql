ALTER TABLE org_settings RENAME COLUMN author_auth_id TO author_auth0_id;
ALTER TABLE users RENAME CONSTRAINT users_auth_id_key TO users_auth0_id_key;
ALTER TABLE users RENAME COLUMN auth_id TO auth0_id;
