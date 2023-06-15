ALTER TABLE IF EXISTS github_app_installs
    ADD COLUMN IF NOT EXISTS url text,
    ADD COLUMN IF NOT EXISTS account_login text,
    ADD COLUMN IF NOT EXISTS account_avatar_url text,
    ADD COLUMN IF NOT EXISTS account_url text,
    ADD COLUMN IF NOT EXISTS account_type text,
    ADD COLUMN IF NOT EXISTS updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW();

CREATE INDEX IF NOT EXISTS github_app_installs_account_login ON github_app_installs (account_login);
