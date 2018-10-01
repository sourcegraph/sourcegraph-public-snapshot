-- Soft-deleted users (deleted_at IS NOT NULL) should not be accounted for in unique indexes.

ALTER TABLE users DROP CONSTRAINT users_username_key;
CREATE UNIQUE INDEX users_username ON users(username) WHERE deleted_at IS NULL;

DROP INDEX users_external_id;
CREATE UNIQUE INDEX users_external_id ON users(external_id, external_provider) WHERE external_provider IS NOT NULL AND deleted_at IS NULL;