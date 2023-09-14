DROP INDEX IF EXISTS github_app_installs_account_login;

ALTER TABLE IF EXISTS github_app_installs
    DROP COLUMN IF EXISTS url,
    DROP COLUMN IF EXISTS account_login,
    DROP COLUMN IF EXISTS account_avatar_url,
    DROP COLUMN IF EXISTS account_url,
    DROP COLUMN IF EXISTS account_type,
    DROP COLUMN IF EXISTS updated_at;
