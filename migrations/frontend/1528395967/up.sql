CREATE TABLE IF NOT EXISTS notebook_stars (
    notebook_id INTEGER NOT NULL REFERENCES notebooks(id) ON DELETE CASCADE DEFERRABLE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE DEFERRABLE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    PRIMARY KEY (notebook_id, user_id)
);

CREATE INDEX IF NOT EXISTS notebook_stars_user_id_idx ON notebook_stars USING btree (user_id);
