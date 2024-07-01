-- delete the constraints
ALTER TABLE IF EXISTS user_credentials DROP CONSTRAINT IF EXISTS check_github_app_id_and_external_service_type_user_credentials;

-- delete the `github_app_id` column
ALTER TABLE IF EXISTS user_credentials DROP COLUMN IF EXISTS github_app_id;
