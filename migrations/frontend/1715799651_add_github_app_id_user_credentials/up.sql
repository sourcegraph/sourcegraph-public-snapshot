-- make the credential column nullable, because we may have user_credentials that are associated with a github_app instead
ALTER TABLE IF EXISTS user_credentials
    ALTER COLUMN credential DROP NOT NULL;

ALTER TABLE IF EXISTS user_credentials
    ADD COLUMN IF NOT EXISTS github_app_id INT NULL REFERENCES github_apps(id);

-- We always one of `credential` or `github_app_id` to be
-- set, but not both. This is a check constraint to ensure that this is always true.
ALTER TABLE IF EXISTS user_credentials
    ADD CONSTRAINT check_credential_and_github_app_id
        CHECK ((credential IS NOT NULL) OR (github_app_id IS NOT NULL));

-- We want to make sure that we never have a user_credential with a `github_app_id` with an `external_service_type`
-- that isn't `github`.
ALTER TABLE IF EXISTS user_credentials
    ADD CONSTRAINT check_github_app_id_and_external_service_type
        CHECK ((github_app_id IS NULL) OR (external_service_type = 'github'));
