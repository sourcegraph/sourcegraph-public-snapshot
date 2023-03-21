CREATE TABLE IF NOT EXISTS user_repo_permissions(
    id SERIAL PRIMARY KEY,
    user_id INT NULL REFERENCES users(id) ON DELETE CASCADE,
    repo_id INT NOT NULL REFERENCES repo(id) ON DELETE CASCADE,
    user_external_account_id INT NULL REFERENCES user_external_accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source TEXT NOT NULL DEFAULT 'sync'
);

CREATE INDEX IF NOT EXISTS user_repo_permissions_user_id_idx ON user_repo_permissions(user_id);
CREATE INDEX IF NOT EXISTS user_repo_permissions_repo_id_idx ON user_repo_permissions(repo_id);
CREATE INDEX IF NOT EXISTS user_repo_permissions_user_external_account_id_idx ON user_repo_permissions(user_external_account_id);
CREATE INDEX IF NOT EXISTS user_repo_permissions_updated_at_idx ON user_repo_permissions(updated_at);
CREATE INDEX IF NOT EXISTS user_repo_permissions_source_idx ON user_repo_permissions(source);
CREATE UNIQUE INDEX IF NOT EXISTS user_repo_permissions_perms_unique_idx ON user_repo_permissions(user_id, user_external_account_id, repo_id);
