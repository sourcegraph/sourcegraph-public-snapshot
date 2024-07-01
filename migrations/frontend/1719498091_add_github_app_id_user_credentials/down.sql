-- delete the constraints
ALTER TABLE IF EXISTS user_credentials DROP CONSTRAINT IF EXISTS check_github_app_id_and_external_service_type_user_credentials;

-- delete the `github_app_id` column
ALTER TABLE IF EXISTS user_credentials DROP COLUMN IF EXISTS github_app_id;

-- delete the constraint replacement
ALTER TABLE IF EXISTS user_credentials DROP CONSTRAINT IF EXISTS user_credentials_github_app_id_fkey_cascade;

-- restore the old constraint
ALTER TABLE ONLY IF EXISTS user_credentials
    ADD CONSTRAINT user_credentials_github_app_id_fkey FOREIGN KEY (github_app_id) REFERENCES github_apps(id);
