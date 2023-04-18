CREATE TABLE github_apps (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL UNIQUE,
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
);-- not adding indexes for now, because we only expect a few rows in the table