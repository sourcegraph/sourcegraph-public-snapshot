CREATE TABLE IF NOT EXISTS github_apps (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL,
    name TEXT NOT NULL,
    slug TEXT NOT NULL,
    base_url TEXT NOT NULL,
    client_id TEXT NOT NULL,
    client_secret TEXT NOT NULL,
    private_key TEXT NOT NULL,
    encryption_key_id TEXT NOT NULL,
    logo TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS github_apps_app_id_slug_base_url_unique
ON github_apps USING btree (app_id, slug, base_url);

