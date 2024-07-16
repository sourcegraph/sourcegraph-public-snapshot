ALTER TABLE IF EXISTS user_credentials
    ADD COLUMN IF NOT EXISTS github_app_id INT NULL REFERENCES github_apps(id) ON DELETE CASCADE;

ALTER TABLE user_credentials DROP CONSTRAINT IF EXISTS check_github_app_id_and_external_service_type_user_credentials;

-- We want to make sure that we never have a user_credential with a `github_app_id` with an `external_service_type`
-- that isn't `github`.
ALTER TABLE IF EXISTS user_credentials
    ADD CONSTRAINT check_github_app_id_and_external_service_type_user_credentials
    CHECK ((github_app_id IS NULL) OR (external_service_type = 'github'));
