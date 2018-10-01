
CREATE TABLE user_external_accounts(
	id SERIAL PRIMARY KEY,
	user_id integer NOT NULL REFERENCES users (id),
	service_type text NOT NULL,
	service_id text NOT NULL,
	account_id text NOT NULL,
	auth_data jsonb,
	account_data jsonb,
	created_at timestamp with time zone NOT NULL DEFAULT now(),
	updated_at timestamp with time zone NOT NULL DEFAULT now(),
	deleted_at timestamp with time zone
);

CREATE UNIQUE INDEX user_external_accounts_account ON user_external_accounts(service_type, service_id, account_id) WHERE deleted_at IS NULL;

-- The users table doesn't record *which* auth provider the users.external_* fields refer to,
-- so we use 'migration_in_progress' for the service_type here and will update it in application
-- code upon next server startup.
INSERT INTO user_external_accounts(user_id, service_type, service_id, account_id)
	SELECT id AS user_id, 'migration_in_progress'::text AS service_type, external_provider AS service_id,
	       external_id AS account_id
	FROM users
	WHERE external_provider IS NOT NULL AND external_id IS NOT NULL;

DROP INDEX users_external_id;
ALTER TABLE users DROP CONSTRAINT check_external_id;
ALTER TABLE users DROP COLUMN external_provider;
ALTER TABLE users DROP COLUMN external_id;

