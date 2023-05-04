CREATE TABLE IF NOT EXISTS github_app_installs (
    id SERIAL PRIMARY KEY,
    app_id INT NOT NULL REFERENCES github_apps(id) ON DELETE CASCADE,
    installation_id INT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS installation_id_idx ON github_app_installs USING btree (installation_id);
CREATE INDEX IF NOT EXISTS app_id_idx ON github_app_installs USING btree (app_id);