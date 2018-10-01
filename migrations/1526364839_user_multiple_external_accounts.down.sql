
ALTER TABLE users ADD COLUMN external_id text;
ALTER TABLE users ADD COLUMN external_provider text;

ALTER TABLE users ADD CONSTRAINT check_external_id CHECK ((external_provider IS NULL) = (external_id IS NULL));
CREATE UNIQUE INDEX users_external_id ON users(external_id, external_provider) WHERE external_provider IS NOT NULL AND deleted_at IS NULL;

UPDATE users SET external_provider=x.service_id, external_id=x.account_id
       FROM user_external_accounts x
       WHERE users.id=x.user_id AND x.deleted_at IS NULL;

DROP TABLE user_external_accounts;

