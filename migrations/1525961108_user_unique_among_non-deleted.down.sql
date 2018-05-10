DROP INDEX users_username;
ALTER TABLE users ADD CONSTRAINT users_username_key UNIQUE(username);

DROP INDEX users_external_id;
CREATE UNIQUE INDEX users_external_id ON users(external_id, external_provider) WHERE external_provider IS NOT NULL;