ALTER TABLE users RENAME COLUMN auth0_id TO auth_id;
ALTER TABLE users RENAME CONSTRAINT users_auth0_id_key TO users_auth_id_key;
ALTER TABLE org_settings RENAME COLUMN author_auth0_id TO author_auth_id;
