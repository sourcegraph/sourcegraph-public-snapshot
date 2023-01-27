BEGIN;

CREATE TABLE IF NOT EXISTS src_permissions(
    id SERIAL PRIMARY KEY,
    user_id INT NULL REFERENCES public.users(id) ON DELETE CASCADE,
    repo_id INT NOT NULL REFERENCES public.repo(id) ON DELETE CASCADE,
    ext_account_id INT NULL REFERENCES public.user_external_accounts(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    source TEXT NOT NULL DEFAULT 'sync'
);

CREATE INDEX IF NOT EXISTS src_permissions_user_id_idx ON src_permissions(user_id);
CREATE INDEX IF NOT EXISTS src_permissions_repo_id_idx ON src_permissions(repo_id);
CREATE INDEX IF NOT EXISTS src_permissions_updated_at_idx ON src_permissions(updated_at);
CREATE INDEX IF NOT EXISTS src_permissions_source_idx ON src_permissions(source);
CREATE UNIQUE INDEX src_permissions_perm_unique_idx ON src_permissions(user_id, repo_id);

END;