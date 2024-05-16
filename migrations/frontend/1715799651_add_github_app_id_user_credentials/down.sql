-- Before removing the github_app_id column from user_credentials, we need to delete all credentials
-- TODO (@BolajiOlajide) figure out the best way to handle null credentials before
-- making the column nullable
DELETE FROM user_credentials WHERE credential IS NULL;

-- delete the constraints
ALTER TABLE IF EXISTS user_credentials DROP CONSTRAINT IF EXISTS check_github_app_id_and_external_service_type;
ALTER TABLE IF EXISTS user_credentials DROP CONSTRAINT IF EXISTS check_credential_and_github_app_id;

-- delete the `github_app_id` column
ALTER TABLE IF EXISTS user_credentials DROP COLUMN IF EXISTS github_app_id;

-- make the `credential` column not nullable
ALTER TABLE IF EXISTS user_credentials ALTER COLUMN credential SET NOT NULL;
